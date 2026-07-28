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

		chartRows, pollErr := s.client.GetSheet(chartSheet)
		if pollErr != nil {
			s.log.WithError(pollErr).Warn("poll chart sheet attempt failed")
			continue
		}

		ohlcv, pollErr = parseColumnarChart(chartRows, s.log)
		if pollErr != nil {
			s.log.WithError(pollErr).Warn("parse columnar chart attempt failed")
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

func parseColumnarChart(chartRows []map[string]interface{}, log *logrus.Logger) ([]OHLCVRow, error) {
	if len(chartRows) == 0 {
		return nil, fmt.Errorf("no data block in response")
	}

	block := chartRows[0]

	rawDate, _ := block["Date"].([]interface{})
	rawOpen, _ := block["Open"].([]interface{})
	rawHigh, _ := block["High"].([]interface{})
	rawLow, _ := block["Low"].([]interface{})
	rawClose, _ := block["Close"].([]interface{})
	rawVolume, _ := block["Volume"].([]interface{})

	// Validate all arrays exist and have same length
	if rawDate == nil {
		return nil, fmt.Errorf("missing Date column in chart response")
	}
	n := len(rawDate)
	if n == 0 {
		return nil, fmt.Errorf("empty Date array in chart response")
	}

	// Check duplicate dates
	dateSet := make(map[string]struct{})
	for _, d := range rawDate {
		if s, ok := d.(string); ok && s != "" {
			dateSet[s] = struct{}{}
		}
	}
	if len(dateSet) > 0 && float64(len(dateSet))/float64(n) < 0.5 {
		return nil, fmt.Errorf("response appears stale: %.0f%% duplicate dates", (1-float64(len(dateSet))/float64(n))*100)
	}

	// Compute median per numeric array for outlier detection
	median := func(arr []interface{}) float64 {
		var vals []float64
		for _, v := range arr {
			f, ok := toFloat64(v)
			if ok {
				vals = append(vals, f)
			}
		}
		if len(vals) == 0 {
			return 0
		}
		// quick median (unsorted, good enough for outlier check)
		var sum float64
		for _, v := range vals {
			sum += v
		}
		return sum / float64(len(vals))
	}

	medianOpen := median(rawOpen)
	medianClose := median(rawClose)

	var result []OHLCVRow
	for i := 0; i < n; i++ {
		date := safeString(rawDate[i])
		if date == "" {
			continue
		}
		if idx := strings.Index(date, "T"); idx > 0 {
			date = date[:idx]
		}

		open := toFloat64Safe(rawOpen, i)
		high := toFloat64Safe(rawHigh, i)
		low := toFloat64Safe(rawLow, i)
		closeV := toFloat64Safe(rawClose, i)
		volume := toInt64Safe(rawVolume, i)

		if open == 0 && high == 0 && low == 0 && closeV == 0 && volume == 0 {
			continue
		}

		// Outlier check: if any value deviates >50% from median, skip this row as unstable
		if medianOpen > 0 && math.Abs(open-medianOpen)/medianOpen > 0.5 {
			log.WithField("index", i).WithField("open", open).WithField("median", medianOpen).Warn("outlier detected in Open, retrying")
			return nil, fmt.Errorf("outlier detected in column Open at index %d: %.2f vs median %.2f", i, open, medianOpen)
		}
		if medianClose > 0 && math.Abs(closeV-medianClose)/medianClose > 0.5 {
			log.WithField("index", i).WithField("close", closeV).WithField("median", medianClose).Warn("outlier detected in Close, retrying")
			return nil, fmt.Errorf("outlier detected in column Close at index %d: %.2f vs median %.2f", i, closeV, medianClose)
		}

		result = append(result, OHLCVRow{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeV,
			Volume: volume,
		})
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("all rows filtered out — data may not be ready yet")
	}

	return result, nil
}

// ponytail: assumes config sheet rows have 2 columns (key, value) regardless of header names
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

func toFloat64Safe(arr []interface{}, i int) float64 {
	if i >= len(arr) {
		return 0
	}
	v := arr[i]
	if v == nil {
		return 0
	}
	s, ok := v.(string)
	if ok {
		if s == "" || s == "#N/A" || s == "N/A" {
			return 0
		}
		f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
		return f
	}
	f, ok := v.(float64)
	if ok {
		return f
	}
	return 0
}

func toInt64Safe(arr []interface{}, i int) int64 {
	if i >= len(arr) {
		return 0
	}
	v := arr[i]
	if v == nil {
		return 0
	}
	s, ok := v.(string)
	if ok {
		if s == "" || s == "#N/A" || s == "N/A" {
			return 0
		}
		f, _ := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
		return int64(f)
	}
	f, ok := v.(float64)
	if ok {
		return int64(f)
	}
	return 0
}

func (s *AnalisisService) runPythonIndicator(ohlcv []OHLCVRow, ticker string) (*Indicators, *SignalResult, *TradingPlan, string, error) {
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

	cmd := exec.CommandContext(ctx, s.pythonBin, s.pythonPath,
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
		Indicators  map[string]interface{} `json:"indicators"`
		Signal      map[string]interface{} `json:"signal"`
		TradingPlan map[string]interface{} `json:"trading_plan"`
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
		EntryZone:                floatPtrVal(m["entry_zone"]),
		EntryPrice:               floatPtrVal(m["entry_price"]),
		StopLoss:                 floatPtrVal(m["stop_loss"]),
		SuggestedPositionSizePct: floatVal(m["suggested_position_size_pct"]),
		SuggestedLots:            intPtrVal(m["suggested_lots"]),
		TimeStopDays:             intVal(m["time_stop_days"]),
		InvalidationNote:         strVal(m["invalidation_note"]),
		Disclaimer:               strVal(m["disclaimer"]),
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
