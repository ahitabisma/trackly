package analisis

type AnalisisRequest struct {
	Ticker    string `json:"ticker" validate:"required"`
	DateStart string `json:"date_start" validate:"required"`
	DateEnd   string `json:"date_end" validate:"required"`
}

type AnalisisResponse struct {
	Error        string             `json:"error,omitempty"`
	Snapshot     *Snapshot          `json:"snapshot,omitempty"`
	OHLCV        []OHLCVRow         `json:"ohlcv,omitempty"`
	Indicators   *Indicators        `json:"indicators,omitempty"`
	Signal       *SignalResult      `json:"signal,omitempty"`
	TradingPlan  *TradingPlan       `json:"trading_plan,omitempty"`
	ChartImage   string             `json:"chart_image,omitempty"`
	AIInsight    string             `json:"ai_insight,omitempty"`
}

type Snapshot struct {
	Kode        string   `json:"kode"`
	CompanyName string   `json:"company_name"`
	Price       float64  `json:"price"`
	High52      float64  `json:"high52"`
	Low52       float64  `json:"low52"`
	Volume      int64    `json:"volume"`
	MarketCap   float64  `json:"marketcap"`
	PE          *float64 `json:"pe"`
	EPS         float64  `json:"eps"`
	Currency    string   `json:"currency"`
}

type OHLCVRow struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

type Indicators struct {
	SMA20        float64  `json:"sma20,omitempty"`
	SMA50        float64  `json:"sma50,omitempty"`
	SMA200       *float64 `json:"sma200,omitempty"`
	EMA20        float64  `json:"ema20,omitempty"`
	EMA50        *float64 `json:"ema50,omitempty"`
	ADX          *float64 `json:"adx,omitempty"`
	DIPlus       *float64 `json:"di_plus,omitempty"`
	DIMinus      *float64 `json:"di_minus,omitempty"`
	RSI          *float64 `json:"rsi,omitempty"`
	MACD         *float64 `json:"macd,omitempty"`
	MACDSignal   *float64 `json:"macd_signal,omitempty"`
	MACDHist     *float64 `json:"macd_histogram,omitempty"`
	StochK       *float64 `json:"stoch_k,omitempty"`
	StochD       *float64 `json:"stoch_d,omitempty"`
	BBUpper      *float64 `json:"bb_upper,omitempty"`
	BBMiddle     *float64 `json:"bb_middle,omitempty"`
	BBLower      *float64 `json:"bb_lower,omitempty"`
	BBWidth      *float64 `json:"bb_width,omitempty"`
	ATR          *float64 `json:"atr,omitempty"`
	OBV          *float64 `json:"obv,omitempty"`
	VolumeMA20   *float64 `json:"volume_ma20,omitempty"`
	VolumeSpike  bool     `json:"volume_spike"`
	Support      float64  `json:"support"`
	Resistance   float64  `json:"resistance"`
	Fib236       *float64 `json:"fib_23_6,omitempty"`
	Fib382       *float64 `json:"fib_38_2,omitempty"`
	Fib500       *float64 `json:"fib_50_0,omitempty"`
	Fib618       *float64 `json:"fib_61_8,omitempty"`
}

type SignalResult struct {
	Overall           string               `json:"overall"`
	Score             float64              `json:"score"`
	Confidence        string               `json:"confidence"`
	TrendFilterPassed *bool                `json:"trend_filter_passed"`
	Breakdown         []SignalBreakdown    `json:"breakdown"`
	Ticker            string               `json:"ticker"`
}

type SignalBreakdown struct {
	Indicator string `json:"indicator"`
	Signal    string `json:"signal"`
	Note      string `json:"note"`
	Score     int    `json:"score"`
}

type EntryZone struct {
	Low  float64 `json:"low"`
	High float64 `json:"high"`
}

type TradingPlan struct {
	Bias                     string        `json:"bias"`
	EntryType                string        `json:"entry_type,omitempty"`
	CurrentPrice             *float64      `json:"current_price,omitempty"`
	EntryPrice               *float64      `json:"entry_price,omitempty"`
	EntryZone                *EntryZone    `json:"entry_zone,omitempty"`
	EntryNote                string        `json:"entry_note,omitempty"`
	StopLoss                 *float64      `json:"stop_loss,omitempty"`
	StopLossBasis            string        `json:"stop_loss_basis,omitempty"`
	Targets                  []TPTarget    `json:"targets,omitempty"`
	SuggestedPositionSizePct float64       `json:"suggested_position_size_pct"`
	SuggestedLots            *int          `json:"suggested_lots"`
	TimeStopDays             int           `json:"time_stop_days"`
	InvalidationNote         string        `json:"invalidation_note"`
	Disclaimer               string        `json:"disclaimer"`
}

type TPTarget struct {
	Level   int     `json:"level"`
	Price   float64 `json:"price"`
	RRRatio float64 `json:"rr_ratio"`
}

type TickerSearchResult struct {
	Kode            string `json:"kode"`
	NamaPerusahaan  string `json:"nama_perusahaan"`
	PapanPencatatan string `json:"papan_pencatatan"`
}

type AiInsightRequest struct {
	Ticker     string      `json:"ticker" validate:"required"`
	DateEnd    string      `json:"date_end" validate:"required"`
	Indicators *Indicators `json:"indicators" validate:"required"`
	Snapshot   *Snapshot   `json:"snapshot"`
}
