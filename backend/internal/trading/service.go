package trading

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"trackly-backend/internal/analisis"
	"trackly-backend/pkg/aiinsight"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var ErrNoTransactions = errors.New("no transactions found for ticker")
var ErrPositionClosed = errors.New("position for ticker is closed")

type TradingService struct {
	db          *gorm.DB
	analisisSvc *analisis.AnalisisService
	log         *logrus.Logger
	pythonBin   string
	pythonPath  string
	nvidiaKey   string
	geminiKey   string
}

func NewTradingService(db *gorm.DB, analisisSvc *analisis.AnalisisService, log *logrus.Logger, pythonPath string, nvidiaKey, geminiKey string) *TradingService {
	return &TradingService{
		db:          db,
		analisisSvc: analisisSvc,
		log:         log,
		pythonBin:   "python",
		pythonPath:  pythonPath,
		nvidiaKey:   nvidiaKey,
		geminiKey:   geminiKey,
	}
}

func (s *TradingService) CreateTransaction(ctx context.Context, req *TransactionRequest, userID uint) (*Transaction, error) {
	var t Transaction
	err := s.db.WithContext(ctx).Raw(
		`INSERT INTO trade_transactions (user_id, ticker, transaction_type, lot, price, transaction_date, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)
		 RETURNING id, user_id, ticker, transaction_type, lot, price, transaction_date, notes, created_at`,
		userID, req.Ticker, req.TransactionType, req.Lot, req.Price, req.TransactionDate, nullableStr(req.Notes),
	).Scan(&t).Error
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}
	return &t, nil
}

func (s *TradingService) GetTransactions(ctx context.Context, userID uint, ticker string) ([]Transaction, error) {
	query := `SELECT id, user_id, ticker, transaction_type, lot, price, transaction_date, notes, created_at
		 FROM trade_transactions WHERE user_id = ? AND deleted_at IS NULL`
	args := []interface{}{userID}
	if ticker != "" {
		query += ` AND ticker = ?`
		args = append(args, ticker)
	}
	query += ` ORDER BY transaction_date ASC, created_at ASC`

	var result []Transaction
	err := s.db.WithContext(ctx).Raw(query, args...).Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	if result == nil {
		result = []Transaction{}
	}
	return result, nil
}

func (s *TradingService) UpdateTransaction(ctx context.Context, id string, userID uint, req *UpdateTransactionRequest) (*Transaction, error) {
	query := `UPDATE trade_transactions SET `
	var args []interface{}
	var sets []string

	if req.Ticker != nil {
		sets = append(sets, "ticker = ?")
		args = append(args, *req.Ticker)
	}
	if req.TransactionType != nil {
		sets = append(sets, "transaction_type = ?")
		args = append(args, *req.TransactionType)
	}
	if req.Lot != nil {
		sets = append(sets, "lot = ?")
		args = append(args, *req.Lot)
	}
	if req.Price != nil {
		sets = append(sets, "price = ?")
		args = append(args, *req.Price)
	}
	if req.TransactionDate != nil {
		sets = append(sets, "transaction_date = ?")
		args = append(args, *req.TransactionDate)
	}
	if req.Notes != nil {
		sets = append(sets, "notes = ?")
		args = append(args, *req.Notes)
	}

	if len(sets) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	for i, s := range sets {
		if i > 0 {
			query += ", "
		}
		query += s
	}
	query += " WHERE id = ? AND user_id = ? AND deleted_at IS NULL"
	args = append(args, id, userID)
	query += " RETURNING id, user_id, ticker, transaction_type, lot, price, transaction_date, notes, created_at"

	var t Transaction
	err := s.db.WithContext(ctx).Raw(query, args...).Scan(&t).Error
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}
	return &t, nil
}

func (s *TradingService) DeleteTransaction(ctx context.Context, id string, userID uint) error {
	err := s.db.WithContext(ctx).Exec(
		`UPDATE trade_transactions SET deleted_at = NOW() WHERE id = ? AND user_id = ? AND deleted_at IS NULL`, id, userID,
	).Error
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

func ComputePosition(transactions []Transaction) Position {
	var runningLot, runningAvg float64

	for _, t := range transactions {
		if t.TransactionType == "buy" {
			totalCost := runningLot*runningAvg + t.Lot*t.Price
			runningLot += t.Lot
			runningAvg = totalCost / runningLot
		} else {
			runningLot -= t.Lot
			if runningLot <= 0 {
				runningLot = 0
				runningAvg = 0
			}
		}
	}

	status := "closed"
	if runningLot > 0 {
		status = "open"
	}

	ticker := ""
	if len(transactions) > 0 {
		ticker = transactions[0].Ticker
	}

	return Position{
		Ticker:          ticker,
		TotalLot:        runningLot,
		AverageBuyPrice: round2(runningAvg),
		Status:          status,
	}
}

func (s *TradingService) GetAllOpenPositions(ctx context.Context, userID uint) ([]Position, error) {
	var tickers []string
	err := s.db.WithContext(ctx).Raw(
		`SELECT DISTINCT ticker FROM trade_transactions WHERE user_id = ? ORDER BY ticker`, userID,
	).Pluck("ticker", &tickers).Error
	if err != nil {
		return nil, fmt.Errorf("query tickers: %w", err)
	}

	var positions []Position
	for _, ticker := range tickers {
		txns, err := s.GetTransactions(ctx, userID, ticker)
		if err != nil {
			s.log.WithError(err).WithField("ticker", ticker).Warn("get transactions failed")
			continue
		}
		pos := ComputePosition(txns)
		if pos.Status == "open" {
			positions = append(positions, pos)
		}
	}
	if positions == nil {
		positions = []Position{}
	}
	return positions, nil
}

func (s *TradingService) GetPositionAnalysis(ctx context.Context, userID uint, ticker string) (*PositionReviewResponse, error) {
	txns, err := s.GetTransactions(ctx, userID, ticker)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	if len(txns) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNoTransactions, ticker)
	}

	position := ComputePosition(txns)
	if position.Status != "open" {
		return nil, fmt.Errorf("%w: %s", ErrPositionClosed, ticker)
	}

	firstBuy := findFirstBuyDate(txns)
	dateStart := firstBuy.AddDate(-1, 0, 0).Format("2006-01-02")
	dateEnd := time.Now().Format("2006-01-02")

	analisisResp, err := s.analisisSvc.RunAnalisis(ctx, &analisis.AnalisisRequest{
		Ticker: ticker, DateStart: dateStart, DateEnd: dateEnd,
	})
	if err != nil {
		return nil, fmt.Errorf("analysis: %w", err)
	}
	if analisisResp.Error != "" {
		return nil, fmt.Errorf("analysis error: %s", analisisResp.Error)
	}

	review, err := s.runPositionReview(analisisResp.OHLCV, ticker, position.AverageBuyPrice, position.TotalLot, firstBuy.Format("2006-01-02"))
	if err != nil {
		s.log.WithError(err).Warn("position review failed, continuing without it")
		review = map[string]interface{}{"error": err.Error()}
	}

	resp := &PositionReviewResponse{
		Ticker:         ticker,
		Position:       position,
		Indicators:     analisisResp.Indicators,
		Signal:         analisisResp.Signal,
		PositionReview: review,
	}

	insightKey := aiinsight.CacheKey(ticker, dateEnd, "position")
	insightData := map[string]interface{}{
		"_cache_key":      insightKey,
		"position":        position,
		"position_review": review,
	}
	if analisisResp.Signal != nil {
		insightData["signal"] = analisisResp.Signal
	}
	insight, err := aiinsight.GenerateInsight(ctx, s.nvidiaKey, s.geminiKey, insightData)
	if err != nil {
		s.log.WithError(err).Warn("ai insight failed")
	} else {
		resp.AIInsight = insight
	}

	return resp, nil
}

func (s *TradingService) runPositionReview(ohlcv []analisis.OHLCVRow, ticker string, buyPrice, lot float64, buyDate string) (map[string]interface{}, error) {
	tmpDir, err := os.MkdirTemp("", fmt.Sprintf("posreview_%s_*", ticker))
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputFile := filepath.Join(tmpDir, "ohlcv.json")
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

	absScript, err := filepath.Abs(s.pythonPath)
	if err != nil {
		return nil, fmt.Errorf("resolve script path: %w", err)
	}
	scriptDir := filepath.Dir(absScript)
	reviewScript := filepath.Join(scriptDir, "run_position_review.py")

	var stdout, stderr bytes.Buffer
	cmd := exec.Command(s.pythonBin, reviewScript,
		"--ohlcv", inputFile,
		"--ticker", ticker,
		"--buy-price", fmt.Sprintf("%.2f", buyPrice),
		"--lot", fmt.Sprintf("%.2f", lot),
		"--buy-date", buyDate,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrStr := strings.TrimSpace(stderr.String())
		if stderrStr != "" {
			return nil, fmt.Errorf("position review python error: %s", stderrStr)
		}
		return nil, fmt.Errorf("position review exec: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("decode position review output: %w", err)
	}
	return result, nil
}

func findFirstBuyDate(txns []Transaction) time.Time {
	earliest := time.Now()
	for _, t := range txns {
		if t.TransactionType != "buy" {
			continue
		}
		d, err := time.Parse("2006-01-02", t.TransactionDate)
		if err != nil {
			continue
		}
		if d.Before(earliest) {
			earliest = d
		}
	}
	return earliest
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}


