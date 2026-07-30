package aiinsight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const nvidiaBaseURL = "https://integrate.api.nvidia.com/v1/chat/completions"
const nvidiaModel = "deepseek-ai/deepseek-v4-flash"

const geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
const geminiModel = "gemini-3.6-flash"

type cacheItem struct {
	content   string
	expiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

var global = &Cache{items: make(map[string]cacheItem)}
var log = logrus.StandardLogger()

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(item.expiresAt) {
		return "", false
	}
	return item.content, true
}

func (c *Cache) Set(key, content string) {
	c.mu.Lock()
	c.items[key] = cacheItem{
		content:   content,
		expiresAt: time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour),
	}
	c.mu.Unlock()
}

func CacheKey(parts ...string) string {
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += ":"
		}
		s += p
	}
	return s
}

const aiInsightSystemPrompt = `Kamu adalah asisten analisis teknikal saham IDX (Bursa Efek Indonesia) yang menulis dalam Bahasa Indonesia untuk investor ritel.

Kamu akan menerima JSON yang BISA salah satu dari dua bentuk:

BENTUK A -- data LENGKAP (ada field "signal" dan/atau "trading_plan" atau "recommendation" dari sistem):
  Tugasmu HANYA menarasikan apa yang sudah dihitung sistem. JANGAN membuat rekomendasi, harga,
  level, atau kesimpulan bias yang BEDA dari yang ada di data -- kamu MENJELASKAN, bukan
  menghitung ulang atau memutuskan sendiri.

BENTUK B -- data MENTAH SAJA (cuma ada "snapshot" harga & "indicators" teknikal, TANPA "signal"
  atau "trading_plan" yang sudah dihitung sistem): Kamu WAJIB melakukan analisis sendiri dari
  indikator yang ada dan menyusun trading plan sendiri, dengan METODOLOGI KONSISTEN berikut (ini
  mengikuti logic yang sama dipakai sistem di tempat lain, supaya hasilnya nggak menyimpang):

  1. TENTUKAN BIAS (bullish/bearish/neutral) dari kombinasi:
     - Trend: SMA20 vs SMA50 (golden/death cross), ADX>=25 dengan arah DI+/DI- kalau trend kuat
     - Momentum: RSI (oversold <30 bullish, overbought >70 bearish, condong ke situ di 30-45/55-70),
       MACD vs signal line, MACD histogram positif/negatif
     - Volatilitas/posisi: harga relatif ke Bollinger Bands (dekat lower band = bullish bias,
       dekat upper band = bearish bias)
     - ADX rendah (<20) = trend lemah, confidence rekomendasi harus lebih rendah/netral
     - Kalau sinyal-sinyal di atas saling bertentangan (mixed), bias = neutral, confidence rendah

  2. KALAU BIAS BEARISH: rekomendasi = HINDARI ENTRY (akun long-only, tidak ada short selling).
     Jangan bikin trading plan (entry/SL/TP) -- cukup jelasin kenapa bearish dan apa yang perlu
     ditunggu supaya berubah jadi layak entry.

  3. KALAU BIAS BULLISH/NEUTRAL, susun trading plan:
     - ENTRY: Bandingkan harga sekarang ke level referensi terdekat (support ATAU sma20, pilih
       yang lebih DEKAT ke harga sekarang). Kalau jarak harga-ke-referensi <= 1.5x ATR, harga
       sekarang masih wajar untuk entry langsung ("market"). Kalau lebih jauh dari itu (harga
       sudah "extended"/naik banyak), JANGAN sarankan entry di harga sekarang -- sarankan tunggu
       pullback ke sekitar level referensi + sedikit buffer (~0.3x ATR di atas referensi), dan
       jelaskan ini sebagai zona entry yang SEMPIT (kisaran +-1.5% dari titik itu, seperti
       "entry di kisaran 140-142", BUKAN rentang lebar dari referensi sampai harga sekarang).
     - STOP LOSS: idealnya di bawah level support dengan buffer kecil (~0.5x ATR). TAPI, kalau
       support yang tersedia jauh sekali di bawah harga (mis. support lama dari beberapa bulan
       lalu), JANGAN pakai jarak itu apa adanya -- batasi risiko maksimum ke sekitar 8% dari
       harga entry. Swing trade butuh risiko yang realistis (~5-8%), BUKAN puluhan persen.
     - TARGET TAKE PROFIT: hitung 3 level dengan rasio risk:reward PERSIS 1:1, 1:2, dan 1:3 dari
       jarak entry-ke-stop-loss (bukan dari ATR terpisah) -- supaya rasionya konsisten dan mudah
       dipahami.
     - Sebutkan juga time stop wajar (mis. sekitar 3-4 minggu / ~20 hari bursa) sebagai batas
       waktu evaluasi ulang kalau plan belum tercapai maupun belum invalid.

  4. Confidence rendah kalau ADX lemah, RSI di zona netral (45-55), atau sinyal-sinyal saling
     bertentangan -- katakan ini secara eksplisit, jangan terdengar lebih yakin dari datanya.

  5. SELALU tutup dengan kalimat: bahwa ini adalah simulasi/analisis informatif berbasis indikator
     teknikal, BUKAN rekomendasi investasi atau ajakan membeli/menjual, dan keputusan akhir ada di
     tangan investor sendiri.

FORMAT OUTPUT (untuk kedua bentuk A dan B), tulis sebagai dua paragraf pendek:
1. INSIGHT: kenapa bias-nya bullish/bearish/netral, pakai bahasa gampang dipahami investor ritel,
   bukan sekadar nyebut ulang label.
2. TRADING PLAN (atau alasan avoid): entry/zona entry, stop loss & alasannya, target TP1-TP3,
   kapan plan ini batal. Kalau ini REVIEW POSISI yang sudah dipegang (ada buy_price/unrealized_pnl),
   mulai dari kondisi posisi (untung/rugi) baru kaitkan ke rekomendasi hold/sell.

ATURAN UMUM:
- Kalau suatu field nggak ada/nggak relevan (mis. bias avoid sehingga tidak ada entry/target),
  jangan mengarang isinya -- lewati saja bagian itu.
- Jangan mengulang semua angka mentah satu-satu kayak baca tabel; rangkai jadi narasi yang
  mengalir dan actionable.
- Total maksimal sekitar 8-10 kalimat untuk dua paragraf itu. Hindari jargon berlebihan tanpa
  penjelasan singkat.`

func callNVIDIA(ctx context.Context, apiKey string, messages []map[string]string, hasPrecomputedSignal bool) (string, error) {
	body := map[string]interface{}{
		"model":       nvidiaModel,
		"messages":    messages,
		"temperature": 0.3,
		"max_tokens":  4000,
	}

	if !hasPrecomputedSignal {
		body["chat_template_kwargs"] = map[string]interface{}{
			"thinking":          true,
			"reasoning_effort": "low",
		}
	}

	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", nvidiaBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	var (
		resp    *http.Response
		lastErr error
	)
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			log.WithField("attempt", attempt+1).Info("Retrying NVIDIA AI insight")
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
		if resp != nil {
			resp.Body.Close()
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			break
		}
		sc := resp.StatusCode
		if sc >= 500 || sc == 429 {
			lastErr = fmt.Errorf("NVIDIA NIM error: HTTP %d (retryable)", sc)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return "", fmt.Errorf("NVIDIA NIM error: HTTP %d - %s", sc, string(body))
	}
	if resp == nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if len(body) > 0 {
				lastErr = fmt.Errorf("%w - %s", lastErr, string(body))
			}
		}
		return "", lastErr
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	content, err := decodeResponseBytes(respBody)
	if err != nil {
		return "", err
	}
	if content == "" {
		log.WithField("raw_response", string(respBody)).Warn("NVIDIA returned empty content")
		return "", fmt.Errorf("empty response")
	}
	return content, nil
}

func callGemini(ctx context.Context, apiKey string, messages []map[string]string) (string, error) {
	body := map[string]interface{}{
		"model":       geminiModel,
		"messages":    messages,
		"temperature": 0.3,
		"max_tokens":  2000,
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", geminiBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini error: HTTP %d - %s", resp.StatusCode, string(respBody))
	}

	content, err := decodeResponseBytes(respBody)
	if err != nil {
		return "", err
	}
	if content == "" {
		log.WithField("raw_response", string(respBody)).Warn("Gemini returned empty content")
		return "", fmt.Errorf("empty response")
	}
	return content, nil
}

func decodeResponseBytes(b []byte) (string, error) {
	var result struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return "", err
	}
	if len(result.Choices) > 0 && result.Choices[0].FinishReason == "length" {
		log.WithFields(logrus.Fields{
			"finish_reason":   "length",
			"content_preview": truncateStr(result.Choices[0].Message.Content, 200),
		}).Warn("AI response truncated by token limit")
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return result.Choices[0].Message.Content, nil
}

func GenerateInsight(ctx context.Context, nvidiaKey, geminiKey string, analysisData map[string]interface{}) (string, error) {
	cacheKey := ""
	if ticker, ok := analysisData["_cache_key"].(string); ok {
		cacheKey = ticker
		delete(analysisData, "_cache_key")
	}
	if cacheKey != "" {
		if cached, ok := global.Get(cacheKey); ok {
			return cached, nil
		}
	}

	hasPrecomputedSignal := false
	if _, ok := analysisData["signal"]; ok {
		hasPrecomputedSignal = true
	}
	if _, ok := analysisData["trading_plan"]; ok {
		hasPrecomputedSignal = true
	}
	if _, ok := analysisData["recommendation"]; ok {
		hasPrecomputedSignal = true
	}

	dataJSON := jsonMarshal(analysisData)
	messages := []map[string]string{
		{"role": "system", "content": aiInsightSystemPrompt},
		{"role": "user", "content": dataJSON},
	}

	if nvidiaKey != "" {
		content, err := callNVIDIA(ctx, nvidiaKey, messages, hasPrecomputedSignal)
		if err == nil {
			if cacheKey != "" {
				global.Set(cacheKey, content)
			}
			return content, nil
		}
		log.WithError(err).Warn("NVIDIA AI insight failed, trying Gemini fallback")
	}

	if geminiKey != "" {
		content, err := callGemini(ctx, geminiKey, messages)
		if err == nil {
			if cacheKey != "" {
				global.Set(cacheKey, content)
			}
			return content, nil
		}
		log.WithError(err).Warn("Gemini AI insight failed")
	}

	return "", fmt.Errorf("layanan ai sedang tidak tersedia")
}

func jsonMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
