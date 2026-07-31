package screening

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"trackly-backend/internal/analisis"
	"trackly-backend/pkg/aiinsight"

	"github.com/sirupsen/logrus"
)

const (
	minAvgVolume            = 1000000.0
	dateRangeDays           = "365"
	pythonTimeout           = 30 * time.Second
	screeningConcurrency    = 3 // ponytail: Apps Script time limit; 3 concurrent fetches stays under 30s client timeout
	screeningMaxRetries     = 2
	screeningRetryBackoff   = 3 * time.Second
)

type Service struct {
	analisisSvc *analisis.AnalisisService
	repo        *Repository
	log         *logrus.Logger
	pythonBin   string
	scriptDir   string
	nvidiaKey   string
	geminiKey   string
}

func NewService(analisisSvc *analisis.AnalisisService, repo *Repository, log *logrus.Logger, scriptDir, nvidiaKey, geminiKey string) *Service {
	absDir, err := filepath.Abs(scriptDir)
	if err != nil {
		absDir = scriptDir
	}
	return &Service{
		analisisSvc: analisisSvc,
		repo:        repo,
		log:         log,
		pythonBin:   "python",
		scriptDir:   absDir,
		nvidiaKey:   nvidiaKey,
		geminiKey:   geminiKey,
	}
}

func (s *Service) findPython() string {
	candidates := []string{"python", "python3", "py"}
	for _, name := range candidates {
		if name == "" {
			continue
		}
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		// verify it's real Python (not Store redirector) by running --version
		cmd := exec.Command(path, "--version")
		if out, err := cmd.CombinedOutput(); err == nil && len(out) > 0 {
			return name
		}
	}
	return s.pythonBin
}

type screeningResult struct {
	Ticker     string  `json:"ticker"`
	Score      float64 `json:"score"`
	Overall    string  `json:"overall"`
	Confidence string  `json:"confidence"`
	AvgVolume  float64 `json:"avg_volume"`
}

func (s *Service) RunNightlyScreening(ctx context.Context) error {
	scanDate := time.Now().In(time.FixedZone("WIB", 7*60*60)).Format("2006-01-02")
	s.log.WithField("scan_date", scanDate).Info("starting nightly screening")

	tickers, err := s.analisisSvc.SearchTickers(ctx)
	if err != nil {
		return fmt.Errorf("fetch tickers: %w", err)
	}

	s.log.WithField("total", len(tickers)).Info("tickers fetched for screening")

	dateEnd := time.Now().Format("2006-01-02")
	dateStart := time.Now().AddDate(0, 0, -365).Format("2006-01-02")

	allTickers := make([]string, len(tickers))
	for i, t := range tickers {
		allTickers[i] = t.Kode
	}

	var screened []screeningResult
	ok, failed := s.runScreeningPass(ctx, allTickers, dateStart, dateEnd)
	screened = append(screened, ok...)

	for attempt := 1; attempt <= screeningMaxRetries && len(failed) > 0; attempt++ {
		time.Sleep(screeningRetryBackoff)
		s.log.WithFields(logrus.Fields{"attempt": attempt, "retry_count": len(failed)}).Info("retrying failed tickers")
		ok, failed = s.runScreeningPass(ctx, failed, dateStart, dateEnd)
		screened = append(screened, ok...)
	}
	if len(failed) > 0 {
		s.log.WithField("still_failing", len(failed)).Warn("screening retries exhausted")
	}

	s.log.WithField("passed", len(screened)).Info("screening pass complete")

	filtered := s.filterAndRank(screened)
	s.log.WithField("top", len(filtered)).Info("filtered top results")

	for i, r := range filtered {
		s.log.WithFields(logrus.Fields{"ticker": r.Ticker, "rank": i + 1}).Info("deep pass")
		plan, aiText, err := s.deepPass(ctx, r.Ticker, dateStart, dateEnd)
		if err != nil {
			s.log.WithField("ticker", r.Ticker).WithError(err).Warn("deep pass failed, saving without plan")
		}

		row := &ScreeningResult{
			ScanDate:    scanDate,
			Ticker:      r.Ticker,
			Rank:        i + 1,
			Score:       math.Round(r.Score*100) / 100,
			Confidence:  r.Confidence,
			Overall:     r.Overall,
			AvgVolume:   r.AvgVolume,
			TradingPlan: plan,
			AIInsight:   aiText,
		}
		if err := s.repo.Upsert(ctx, row); err != nil {
			s.log.WithField("ticker", r.Ticker).WithError(err).Error("upsert failed")
		}
	}

	s.log.WithField("scan_date", scanDate).Info("nightly screening complete")
	return nil
}

func (s *Service) screeningPass(ctx context.Context, ticker, dateStart, dateEnd string) (*screeningResult, error) {
	s.log.WithField("ticker", ticker).Info("screening ticker")
	ohlcv, err := s.analisisSvc.FetchOHLCV(ctx, ticker, dateStart, dateEnd)
	if err != nil {
		return nil, fmt.Errorf("fetch ohlcv: %w", err)
	}
	if len(ohlcv) < 30 {
		return nil, fmt.Errorf("insufficient data: %d bars", len(ohlcv))
	}

	result, err := s.runScreeningPython(ohlcv, ticker)
	if err != nil {
		return nil, fmt.Errorf("python screening: %w", err)
	}

	return result, nil
}

type tickerResult struct {
	ticker string
	result *screeningResult
	err    error
}

func (s *Service) runScreeningPass(ctx context.Context, tickers []string, dateStart, dateEnd string) ([]screeningResult, []string) {
	results := make(chan tickerResult, len(tickers))
	sem := make(chan struct{}, screeningConcurrency)

	for _, kode := range tickers {
		go func(k string) {
			sem <- struct{}{}
			defer func() { <-sem }()
			r, err := s.screeningPass(ctx, k, dateStart, dateEnd)
			results <- tickerResult{ticker: k, result: r, err: err}
		}(kode)
	}

	var ok []screeningResult
	var failed []string
	for range tickers {
		tr := <-results
		if tr.err != nil {
			s.log.WithField("ticker", tr.ticker).WithError(tr.err).Warn("screening pass skipped")
			failed = append(failed, tr.ticker)
			continue
		}
		if tr.result != nil {
			ok = append(ok, *tr.result)
		}
	}
	return ok, failed
}

type screeningPyOutput struct {
	Error             string  `json:"error"`
	Score             float64 `json:"score"`
	Overall           string  `json:"overall"`
	Confidence        string  `json:"confidence"`
	AvgVolume         float64 `json:"avg_volume"`
	TrendFilterPassed bool    `json:"trend_filter_passed"`
}

func (s *Service) runScreeningPython(ohlcv []analisis.OHLCVRow, ticker string) (*screeningResult, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("screening_%s_*", ticker))
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := make([]map[string]interface{}, len(ohlcv))
	for i, row := range ohlcv {
		inputData[i] = map[string]interface{}{
			"date": row.Date, "open": row.Open, "high": row.High,
			"low": row.Low, "close": row.Close, "volume": row.Volume,
		}
	}
	f, err := os.Create(inputFile)
	if err != nil {
		return nil, fmt.Errorf("create input file: %w", err)
	}
	if err := json.NewEncoder(f).Encode(inputData); err != nil {
		f.Close()
		return nil, fmt.Errorf("write input: %w", err)
	}
	f.Close()

	script := filepath.Join(s.scriptDir, "screening_pass.py")
	ctx, cancel := context.WithTimeout(context.Background(), pythonTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, s.findPython(), script, "--input", inputFile, "--ticker", ticker)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("python error: %s", stderrStr)
		}
		return nil, fmt.Errorf("python exec: %w", err)
	}

	var py screeningPyOutput
	if err := json.Unmarshal(stdout.Bytes(), &py); err != nil {
		return nil, fmt.Errorf("decode python output: %w", err)
	}
	if py.Error != "" {
		return nil, fmt.Errorf("python error: %s", py.Error)
	}

	return &screeningResult{
		Ticker:     ticker,
		Score:      py.Score,
		Overall:    py.Overall,
		Confidence: py.Confidence,
		AvgVolume:  py.AvgVolume,
	}, nil
}

func (s *Service) filterAndRank(results []screeningResult) []screeningResult {
	var filtered []screeningResult
	for _, r := range results {
		if r.AvgVolume < minAvgVolume {
			continue
		}
		if r.Overall == "bearish" {
			continue
		}
		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	if len(filtered) > 10 {
		filtered = filtered[:10]
	}
	return filtered
}

type deepPyOutput struct {
	Error       string                 `json:"error"`
	Indicators  map[string]interface{} `json:"indicators"`
	Signal      map[string]interface{} `json:"signal"`
	TradingPlan map[string]interface{} `json:"trading_plan"`
}

func (s *Service) runDeepPython(ohlcv []analisis.OHLCVRow, ticker string) (map[string]interface{}, map[string]interface{}, map[string]interface{}, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("deep_%s_*", ticker))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.json")
	inputData := make([]map[string]interface{}, len(ohlcv))
	for i, row := range ohlcv {
		inputData[i] = map[string]interface{}{
			"date": row.Date, "open": row.Open, "high": row.High,
			"low": row.Low, "close": row.Close, "volume": row.Volume,
		}
	}
	f, err := os.Create(inputFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create input file: %w", err)
	}
	if err := json.NewEncoder(f).Encode(inputData); err != nil {
		f.Close()
		return nil, nil, nil, fmt.Errorf("write input: %w", err)
	}
	f.Close()

	script := filepath.Join(s.scriptDir, "deep_pass.py")
	ctx, cancel := context.WithTimeout(context.Background(), pythonTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, s.findPython(), script, "--input", inputFile, "--ticker", ticker)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, nil, nil, fmt.Errorf("python error: %s", stderrStr)
		}
		return nil, nil, nil, fmt.Errorf("python exec: %w", err)
	}

	var py deepPyOutput
	if err := json.Unmarshal(stdout.Bytes(), &py); err != nil {
		return nil, nil, nil, fmt.Errorf("decode python output: %w", err)
	}
	if py.Error != "" {
		return nil, nil, nil, fmt.Errorf("python error: %s", py.Error)
	}
	if py.Indicators == nil || py.Signal == nil || py.TradingPlan == nil {
		return nil, nil, nil, fmt.Errorf("incomplete python output")
	}
	return py.Indicators, py.Signal, py.TradingPlan, nil
}

func (s *Service) deepPass(ctx context.Context, ticker, dateStart, dateEnd string) (tradingPlanJSON string, aiText string, err error) {
	ohlcv, err := s.analisisSvc.FetchOHLCV(ctx, ticker, dateStart, dateEnd)
	if err != nil {
		return "", "", fmt.Errorf("fetch ohlcv: %w", err)
	}
	if len(ohlcv) < 30 {
		return "", "", fmt.Errorf("insufficient data: %d bars", len(ohlcv))
	}

	indicators, signal, plan, err := s.runDeepPython(ohlcv, ticker)
	if err != nil {
		return "", "", fmt.Errorf("deep python: %w", err)
	}

	planJSON, _ := json.Marshal(plan)
	planStr := string(planJSON)

	insightData := map[string]interface{}{
		"ticker":        ticker,
		"indicators":    indicators,
		"signal":        signal,
		"trading_plan":  plan,
	}

	insight, err := aiinsight.GenerateInsight(ctx, s.nvidiaKey, s.geminiKey, insightData)
	if err != nil {
		s.log.WithField("ticker", ticker).WithError(err).Warn("AI insight failed")
		insight = ""
	}

	return planStr, insight, nil
}
