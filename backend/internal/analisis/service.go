package analisis

import (
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
	pythonPath     string
	pollInterval   time.Duration
	pollMaxRetries int
}

func NewAnalisisService(client *appsscript.Client, log *logrus.Logger, pythonPath string, pollInterval time.Duration, pollMaxRetries int) *AnalisisService {
	return &AnalisisService{
		client:         client,
		log:            log,
		pythonPath:     pythonPath,
		pollInterval:   pollInterval,
		pollMaxRetries: pollMaxRetries,
	}
}

type masterCache struct {
	data      []TickerSearchResult
	expiresAt time.Time
}

var (
	masterCacheData  *masterCache
	masterCacheMu    sync.RWMutex
	masterCacheTTL   = 5 * time.Minute
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
	if err := s.client.SetValueByKey("selected_ticker", req.Ticker); err != nil {
		return nil, fmt.Errorf("update selected_ticker: %w", err)
	}
	if err := s.client.SetValueByKey("date_start", req.DateStart); err != nil {
		return nil, fmt.Errorf("update date_start: %w", err)
	}
	if err := s.client.SetValueByKey("date_end", req.DateEnd); err != nil {
		return nil, fmt.Errorf("update date_end: %w", err)
	}

	var chartRows []map[string]interface{}
	var pollErr error
	for attempt := 0; attempt <= s.pollMaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(s.pollInterval):
		}

		chartRows, pollErr = s.client.GetSheet("chart")
		if pollErr != nil {
			s.log.WithError(pollErr).Warn("poll chart sheet attempt failed")
			continue
		}

		if len(chartRows) > 0 {
			break
		}
	}
	if pollErr != nil && len(chartRows) == 0 {
		return nil, fmt.Errorf("chart sheet not available after polling: %w", pollErr)
	}

	var ohlcv []OHLCVRow
	for _, row := range chartRows {
		date := strings.TrimSpace(safeString(row["Date"]))
		if date == "" || date == "#N/A" {
			continue
		}
		// ponytail: strip ISO time suffix, keep date portion
		if idx := strings.Index(date, "T"); idx > 0 {
			date = date[:idx]
		}
		ohlcv = append(ohlcv, OHLCVRow{
			Date:   date,
			Open:   safeFloat(row["Open"]),
			High:   safeFloat(row["High"]),
			Low:    safeFloat(row["Low"]),
			Close:  safeFloat(row["Close"]),
			Volume: safeInt64(row["Volume"]),
		})
	}

	snapshot, _ := s.GetTicker(ctx, req.Ticker)

	resp := &AnalisisResponse{
		Snapshot: snapshot,
		OHLCV:    ohlcv,
	}
	if len(ohlcv) > 0 {
		ind, sig, plan, chartImg, err := s.runPythonIndicator(ohlcv, req.Ticker)
		if err != nil {
			s.log.WithError(err).Warn("python indicator calculation failed, skipping")
		} else {
			resp.Indicators = ind
			resp.Signal = sig
			resp.TradingPlan = plan
			resp.ChartImage = chartImg
		}
	}
	return resp, nil
}

func (s *AnalisisService) runPythonIndicator(ohlcv []OHLCVRow, ticker string) (*Indicators, *SignalResult, *TradingPlan, string, error) {
	tmpDir := os.TempDir()
	inputFile := filepath.Join(tmpDir, fmt.Sprintf("chart_%s_%d.json", ticker, time.Now().UnixNano()))
	outFile := filepath.Join(tmpDir, fmt.Sprintf("chart_%s_%d.png", ticker, time.Now().UnixNano()))
	defer os.Remove(inputFile)
	defer os.Remove(outFile)

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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pythonCmd := "python"
	// ponytail: use "python3" on linux, configurable if needed
	cmd := exec.CommandContext(ctx, pythonCmd, s.pythonPath,
		"--input", inputFile,
		"--ticker", ticker,
		"--out", outFile,
	)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, nil, nil, "", fmt.Errorf("python error: %s", string(exitErr.Stderr))
		}
		return nil, nil, nil, "", fmt.Errorf("python exec: %w", err)
	}

	type pyOutput struct {
		Indicators   map[string]interface{} `json:"indicators"`
		Signal       map[string]interface{} `json:"signal"`
		TradingPlan  map[string]interface{} `json:"trading_plan"`
	}
	var py pyOutput
	if err := json.Unmarshal(output, &py); err != nil {
		return nil, nil, nil, "", fmt.Errorf("decode python output: %w", err)
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
		Overall: strVal(m["overall"]),
		Score:   floatVal(m["score"]),
		Ticker:  strVal(m["ticker"]),
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
		Bias:                    strVal(m["bias"]),
		EntryZone:               floatPtrVal(m["entry_zone"]),
		EntryPrice:              floatPtrVal(m["entry_price"]),
		StopLoss:                floatPtrVal(m["stop_loss"]),
		SuggestedPositionSizePct: floatVal(m["suggested_position_size_pct"]),
		InvalidationNote:        strVal(m["invalidation_note"]),
		Disclaimer:              strVal(m["disclaimer"]),
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
