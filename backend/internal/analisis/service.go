package analisis

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"trackly-backend/pkg/appsscript"

	"github.com/sirupsen/logrus"
)

type AnalisisService struct {
	client         *appsscript.Client
	log            *logrus.Logger
	pool           *WorkerPool
	pythonPath     string
	pythonBin      string
	pythonTimeout  time.Duration
	pollInterval   time.Duration
	pollMaxRetries int
}

func NewAnalisisService(client *appsscript.Client, log *logrus.Logger, pool *WorkerPool, pythonPath string, pollInterval time.Duration, pollMaxRetries int, opts ...func(*AnalisisService)) *AnalisisService {
	s := &AnalisisService{
		client:         client,
		log:            log,
		pool:           pool,
		pythonPath:     pythonPath,
		pythonBin:      "python",
		pythonTimeout:  120 * time.Second,
		pollInterval:   pollInterval,
		pollMaxRetries: pollMaxRetries,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithPythonBin(bin string) func(*AnalisisService) {
	return func(s *AnalisisService) { s.pythonBin = bin }
}

func WithPythonTimeout(timeout time.Duration) func(*AnalisisService) {
	return func(s *AnalisisService) { s.pythonTimeout = timeout }
}

type masterCache struct {
	data      []TickerSearchResult
	expiresAt time.Time
}

var (
	masterCacheData *masterCache
	masterCacheMu   sync.RWMutex
	masterCacheTTL  = 5 * time.Minute
)

func safeString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprint(val)
	}
}

func safeFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case string:
		s := strings.TrimSpace(val)
		if s == "" || s == "#N/A" || s == "N/A" {
			return 0
		}
		f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
		return f
	default:
		return 0
	}
}

func safeInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int64(val)
	case string:
		s := strings.TrimSpace(val)
		if s == "" || s == "#N/A" || s == "N/A" {
			return 0
		}
		f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
		return int64(f)
	default:
		return 0
	}
}

func safeFloatPtr(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case float64:
		return &val
	case string:
		s := strings.TrimSpace(val)
		if s == "" || s == "#N/A" || s == "N/A" || s == "0" {
			return nil
		}
		f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
		return &f
	default:
		return nil
	}
}

func (s *AnalisisService) SearchTickers(ctx context.Context) ([]TickerSearchResult, error) {
	masterCacheMu.RLock()
	cached := masterCacheData
	masterCacheMu.RUnlock()

	if cached == nil || time.Now().After(cached.expiresAt) {
		rows, err := s.client.GetSheet("master")
		if err != nil {
			if cached != nil {
				s.log.WithError(err).Warn("failed to refresh master cache, using stale")
			} else {
				return nil, fmt.Errorf("fetch master sheet: %w", err)
			}
		} else {
			var results []TickerSearchResult
			for _, row := range rows {
				results = append(results, TickerSearchResult{
					Kode:            safeString(row["Kode"]),
					NamaPerusahaan:  safeString(row["Nama Perusahaan"]),
					PapanPencatatan: safeString(row["Papan Pencatatan"]),
				})
			}
			masterCacheMu.Lock()
			masterCacheData = &masterCache{
				data:      results,
				expiresAt: time.Now().Add(masterCacheTTL),
			}
			masterCacheMu.Unlock()
			cached = masterCacheData
		}
	}

	return cached.data, nil
}

func (s *AnalisisService) GetTicker(ctx context.Context, kode string) (*Snapshot, error) {
	rows, err := s.client.GetSheet("google_finance")
	if err != nil {
		return nil, fmt.Errorf("fetch google_finance sheet: %w", err)
	}

	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(safeString(row["ticker"])), kode) {
			return &Snapshot{
				Kode:        strings.TrimSpace(safeString(row["ticker"])),
				CompanyName: strings.TrimSpace(safeString(row["company_name"])),
				Price:       safeFloat(row["price"]),
				High52:      safeFloat(row["high52"]),
				Low52:       safeFloat(row["low52"]),
				Volume:      safeInt64(row["volume"]),
				MarketCap:   safeFloat(row["marketcap"]),
				PE:          safeFloatPtr(row["pe"]),
				EPS:         safeFloat(row["eps"]),
				Currency:    strings.TrimSpace(safeString(row["currency"])),
			}, nil
		}
	}

	return nil, fmt.Errorf("ticker %s not found in google_finance", kode)
}

func (s *AnalisisService) RunAnalisis(ctx context.Context, req *AnalisisRequest) (*AnalisisResponse, error) {
	slot, err := s.pool.AcquireWorker(ctx)
	if err != nil {
		return nil, err
	}

	var released bool
	release := func() {
		if !released {
			released = true
			s.pool.ReleaseWorker(slot)
		}
	}
	defer release()

	acquiredAt := time.Now()
	s.pool.Watchdog(slot, acquiredAt)

	// ponytail: echo-check max 3 retries with 500ms interval
	echoCheck := func() bool {
		for attempt := 0; attempt < 3; attempt++ {
			rows, err := s.client.GetSheet("config")
			if err != nil {
				s.log.WithError(err).Warn("echo-check: get config failed")
				time.Sleep(500 * time.Millisecond)
				continue
			}
			m := buildConfigMap(rows)
			if v, ok := m[fmt.Sprintf("selected_ticker_%d", slot)]; ok && v == req.Ticker {
				return true
			}
			s.log.WithField("config", m).Warn("echo-check: ticker mismatch, retrying")
			time.Sleep(500 * time.Millisecond)
		}
		return false
	}

	if err := s.client.SetValueByKey(fmt.Sprintf("selected_ticker_%d", slot), req.Ticker); err != nil {
		return nil, fmt.Errorf("update selected_ticker_%d: %w", slot, err)
	}
	if err := s.client.SetValueByKey(fmt.Sprintf("date_start_%d", slot), req.DateStart); err != nil {
		return nil, fmt.Errorf("update date_start_%d: %w", slot, err)
	}
	if err := s.client.SetValueByKey(fmt.Sprintf("date_end_%d", slot), req.DateEnd); err != nil {
		return nil, fmt.Errorf("update date_end_%d: %w", slot, err)
	}

	if !echoCheck() {
		rows, _ := s.client.GetSheet("config")
		cfgDump := buildConfigMap(rows)
		return nil, fmt.Errorf("echo-check failed: slot %d ticker %q not in config: %v", slot, req.Ticker, cfgDump)
	}

	chartSheet := fmt.Sprintf("chart_%d", slot)
	var ohlcv []OHLCVRow
	var pollErr error

	for attempt := 0; attempt <= s.pollMaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.pollInterval):
		}

		var chartRows []map[string]interface{}
		chartRows, pollErr = s.client.GetSheet(chartSheet)
		if pollErr != nil {
			s.log.WithError(pollErr).Warn("poll chart sheet attempt failed")
			continue
		}

		ohlcv, pollErr = parseChartRows(chartRows)
		if pollErr != nil {
			s.log.WithError(pollErr).Warn("parse chart rows failed")
			continue
		}

		if len(ohlcv) > 0 {
			break
		}
	}

	if pollErr != nil && len(ohlcv) == 0 {
		return nil, fmt.Errorf("chart sheet not available after polling: %w", pollErr)
	}

	// Release worker early — chart data is safe, python is async from pool
	release()

	snapshot, err := s.GetTicker(ctx, req.Ticker)
	if err != nil {
		s.log.WithError(err).Warn("snapshot data unavailable, continuing without it")
	}

	resp := &AnalisisResponse{
		Snapshot: snapshot,
		OHLCV:    ohlcv,
	}
	if len(ohlcv) > 0 {
		ind, sig, plan, chartImg, err := s.runPythonIndicator(ohlcv, req.Ticker)
		if err != nil {
			resp.Error = fmt.Sprintf("indicator calculation failed: %v", err)
			s.log.Warnf("python indicator calculation failed: %v", err)
		} else {
			resp.Indicators = ind
			resp.Signal = sig
			resp.TradingPlan = plan
			resp.ChartImage = chartImg
		}
	}
	return resp, nil
}

func parseChartRows(chartRows []map[string]interface{}) ([]OHLCVRow, error) {
	if len(chartRows) == 0 {
		return nil, fmt.Errorf("empty chart data")
	}

	var result []OHLCVRow
	dateSet := make(map[string]struct{})

	for _, row := range chartRows {
		date := safeString(row["Date"])
		if date == "" {
			continue
		}
		if idx := strings.Index(date, "T"); idx > 0 {
			date = date[:idx]
		}

		open := safeFloat(row["Open"])
		high := safeFloat(row["High"])
		low := safeFloat(row["Low"])
		closeV := safeFloat(row["Close"])
		volume := safeInt64(row["Volume"])

		if open == 0 || high == 0 || low == 0 || closeV == 0 {
			continue
		}

		dateSet[date] = struct{}{}
		result = append(result, OHLCVRow{
			Date: date, Open: open, High: high,
			Low: low, Close: closeV, Volume: volume,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("all rows filtered out — data not ready yet")
	}

	uniqRatio := float64(len(dateSet)) / float64(len(result))
	if uniqRatio < 0.2 {
		return nil, fmt.Errorf("stale data: %.0f%% dates are duplicates", (1-uniqRatio)*100)
	}

	return result, nil
}

func buildConfigMap(rows []map[string]interface{}) map[string]string {
	m := make(map[string]string, len(rows))
	for _, row := range rows {
		var key, val string
		for _, v := range row {
			s := safeString(v)
			if strings.HasPrefix(s, "selected_ticker_") ||
				strings.HasPrefix(s, "date_start_") ||
				strings.HasPrefix(s, "date_end_") {
				key = s
			} else if s != "" {
				val = s
			}
		}
		if key != "" {
			m[key] = val
		}
	}
	return m
}

func (s *AnalisisService) findPython() string {
	candidates := []string{"python3", "py", "python"}
	if s.pythonBin != "python" {
		candidates = append([]string{s.pythonBin}, candidates...)
	}
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

func (s *AnalisisService) runPythonIndicator(ohlcv []OHLCVRow, ticker string) (*Indicators, *SignalResult, *TradingPlan, string, error) {
	pythonBin := s.findPython()
	absScript, err := filepath.Abs(s.pythonPath)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("resolve python script path: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"python_bin":   pythonBin,
		"script_path":  absScript,
		"ticker":       ticker,
		"ohlcv_bars":   len(ohlcv),
	}).Debug("running python indicator script")

	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("trackly_%s_*", ticker))
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "input.json")
	outFile := filepath.Join(tmpDir, "chart.png")

	inputData := make([]map[string]interface{}, len(ohlcv))
	for i, row := range ohlcv {
		inputData[i] = map[string]interface{}{
			"date":   row.Date,
			"open":   row.Open,
			"high":   row.High,
			"low":    row.Low,
			"close":  row.Close,
			"volume": row.Volume,
		}
	}

	f, err := os.Create(inputFile)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("create temp file: %w", err)
	}
	if err := json.NewEncoder(f).Encode(inputData); err != nil {
		f.Close()
		return nil, nil, nil, "", fmt.Errorf("write temp file: %w", err)
	}
	f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), s.pythonTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, pythonBin, absScript,
		"--input", inputFile,
		"--ticker", ticker,
		"--out", outFile,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		stdoutStr := strings.TrimSpace(stdout.String())
		s.log.WithFields(logrus.Fields{
			"python_bin":  pythonBin,
			"script_path": absScript,
			"stdout":      stdoutStr,
			"stderr":      stderrStr,
			"stdout_len":  len(stdoutStr),
			"stderr_len":  len(stderrStr),
		}).Warn("python process failed")
		if stderrStr != "" {
			return nil, nil, nil, "", fmt.Errorf("python error (bin=%s script=%s): %s", pythonBin, absScript, stderrStr)
		}
		return nil, nil, nil, "", fmt.Errorf("python exec (bin=%s): %w", pythonBin, err)
	}

	stdoutBytes := stdout.Bytes()
	if len(bytes.TrimSpace(stdoutBytes)) == 0 {
		stderrStr := strings.TrimSpace(stderr.String())
		s.log.WithFields(logrus.Fields{
			"python_bin":  pythonBin,
			"script_path": absScript,
			"stderr":      stderrStr,
			"stderr_len":  len(stderrStr),
		}).Warn("python produced empty stdout")
		return nil, nil, nil, "", fmt.Errorf("python produced empty stdout (stderr: %s)", stderrStr)
	}

	type pyOutput struct {
		Error       string                 `json:"error"`
		Indicators  map[string]interface{} `json:"indicators"`
		Signal      map[string]interface{} `json:"signal"`
		TradingPlan map[string]interface{} `json:"trading_plan"`
	}
	var py pyOutput
	if err := json.Unmarshal(stdoutBytes, &py); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		s.log.WithFields(logrus.Fields{
			"stdout":      string(stdoutBytes),
			"stdout_len":  len(stdoutBytes),
			"stderr":      stderrStr,
			"stderr_len":  len(stderrStr),
		}).Warn("python output decode failed")
		return nil, nil, nil, "", fmt.Errorf("decode python output (stdout=%q): %w", string(stdoutBytes), err)
	}
	if py.Error != "" {
		return nil, nil, nil, "", fmt.Errorf("python error: %s", py.Error)
	}
	if py.Indicators == nil || py.Signal == nil || py.TradingPlan == nil {
		return nil, nil, nil, "", fmt.Errorf("incomplete python output: indicators=%v signal=%v trading_plan=%v", py.Indicators == nil, py.Signal == nil, py.TradingPlan == nil)
	}

	ind := mapIndicators(py.Indicators)
	sig := mapSignal(py.Signal)
	plan := mapTradingPlan(py.TradingPlan)

	chartBase64 := ""
	if b, err := os.ReadFile(outFile); err == nil {
		chartBase64 = base64.StdEncoding.EncodeToString(b)
	}

	return ind, sig, plan, chartBase64, nil
}

func mapIndicators(m map[string]interface{}) *Indicators {
	if m == nil {
		return nil
	}
	return &Indicators{
		SMA20:       floatVal(m["sma20"]),
		SMA50:       floatVal(m["sma50"]),
		SMA200:      floatPtrVal(m["sma200"]),
		EMA20:       floatVal(m["ema20"]),
		EMA50:       floatPtrVal(m["ema50"]),
		ADX:         floatPtrVal(m["adx"]),
		DIPlus:      floatPtrVal(m["di_plus"]),
		DIMinus:     floatPtrVal(m["di_minus"]),
		RSI:         floatPtrVal(m["rsi"]),
		MACD:        floatPtrVal(m["macd"]),
		MACDSignal:  floatPtrVal(m["macd_signal"]),
		MACDHist:    floatPtrVal(m["macd_histogram"]),
		StochK:      floatPtrVal(m["stoch_k"]),
		StochD:      floatPtrVal(m["stoch_d"]),
		BBUpper:     floatPtrVal(m["bb_upper"]),
		BBMiddle:    floatPtrVal(m["bb_middle"]),
		BBLower:     floatPtrVal(m["bb_lower"]),
		BBWidth:     floatPtrVal(m["bb_width"]),
		ATR:         floatPtrVal(m["atr"]),
		OBV:         floatPtrVal(m["obv"]),
		VolumeMA20:  floatPtrVal(m["volume_ma20"]),
		VolumeSpike: boolVal(m["volume_spike"]),
		Support:     floatVal(m["support"]),
		Resistance:  floatVal(m["resistance"]),
		Fib236:      floatPtrVal(m["fib_23_6"]),
		Fib382:      floatPtrVal(m["fib_38_2"]),
		Fib500:      floatPtrVal(m["fib_50_0"]),
		Fib618:      floatPtrVal(m["fib_61_8"]),
	}
}

func mapSignal(m map[string]interface{}) *SignalResult {
	if m == nil {
		return nil
	}
	sig := &SignalResult{
		Overall:           strVal(m["overall"]),
		Score:             floatVal(m["score"]),
		Confidence:        strVal(m["confidence"]),
		TrendFilterPassed: boolPtrVal(m["trend_filter_passed"]),
		Ticker:            strVal(m["ticker"]),
	}
	if bd, ok := m["breakdown"].([]interface{}); ok {
		for _, item := range bd {
			if bm, ok := item.(map[string]interface{}); ok {
				sig.Breakdown = append(sig.Breakdown, SignalBreakdown{
					Indicator: strVal(bm["indicator"]),
					Signal:    strVal(bm["signal"]),
					Note:      strVal(bm["note"]),
					Score:     intVal(bm["score"]),
				})
			}
		}
	}
	return sig
}

func mapTradingPlan(m map[string]interface{}) *TradingPlan {
	if m == nil {
		return nil
	}
	plan := &TradingPlan{
		Bias:                     strVal(m["bias"]),
		EntryType:                strVal(m["entry_type"]),
		CurrentPrice:             floatPtrVal(m["current_price"]),
		EntryPrice:               floatPtrVal(m["entry_price"]),
		EntryNote:                strVal(m["entry_note"]),
		StopLoss:                 floatPtrVal(m["stop_loss"]),
		StopLossBasis:            strVal(m["stop_loss_basis"]),
		SuggestedPositionSizePct: floatVal(m["suggested_position_size_pct"]),
		SuggestedLots:            intPtrVal(m["suggested_lots"]),
		TimeStopDays:             intVal(m["time_stop_days"]),
		InvalidationNote:         strVal(m["invalidation_note"]),
		Disclaimer:               strVal(m["disclaimer"]),
	}
	if ez, ok := m["entry_zone"].(map[string]interface{}); ok {
		plan.EntryZone = &EntryZone{
			Low:  floatVal(ez["low"]),
			High: floatVal(ez["high"]),
		}
	}
	if tgts, ok := m["targets"].([]interface{}); ok {
		for _, item := range tgts {
			if tm, ok := item.(map[string]interface{}); ok {
				plan.Targets = append(plan.Targets, TPTarget{
					Level:   intVal(tm["level"]),
					Price:   floatVal(tm["price"]),
					RRRatio: floatVal(tm["rr_ratio"]),
				})
			}
		}
	}
	return plan
}

func floatVal(v interface{}) float64 {
	if v == nil {
		return 0
	}
	f, _ := toFloat64(v)
	return f
}

func floatPtrVal(v interface{}) *float64 {
	if v == nil {
		return nil
	}
	f, ok := toFloat64(v)
	if !ok {
		return nil
	}
	return &f
}

func intVal(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	default:
		return 0
	}
}

func strVal(v interface{}) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func boolVal(v interface{}) bool {
	b, _ := v.(bool)
	return b
}

func boolPtrVal(v interface{}) *bool {
	if v == nil {
		return nil
	}
	b, ok := v.(bool)
	if !ok {
		return nil
	}
	return &b
}

func intPtrVal(v interface{}) *int {
	if v == nil {
		return nil
	}
	f, ok := toFloat64(v)
	if !ok {
		return nil
	}
	i := int(f)
	return &i
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return math.Round(val*100) / 100, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
