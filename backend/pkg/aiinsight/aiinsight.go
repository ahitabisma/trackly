package aiinsight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const nvidiaBaseURL = "https://integrate.api.nvidia.com/v1/chat/completions"
const nvidiaModel = "deepseek-ai/deepseek-v4-flash"

type cacheItem struct {
	content   string
	expiresAt time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
}

var global = &Cache{items: make(map[string]cacheItem)}

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

func GenerateInsight(ctx context.Context, apiKey string, analysisData map[string]interface{}) (string, error) {
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

	dataJSON, _ := json.Marshal(analysisData)

	if apiKey == "" {
		return "", fmt.Errorf("NVIDIA_API_KEY not set")
	}

	reqBody := map[string]interface{}{
		"model": nvidiaModel,
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "Kamu membuat ringkasan naratif 2-4 kalimat Bahasa " +
					"Indonesia dari data analisis saham yang diberikan. JANGAN " +
					"membuat rekomendasi baru yang berbeda dari data yang ada -- " +
					"tugasmu MENJELASKAN angka & kesimpulan yang sudah dihitung " +
					"sistem, bukan membuat keputusan independen. Jangan mengulang " +
					"semua angka mentah, fokus ke insight yang membantu pembacaan.",
			},
			{"role": "user", "content": string(dataJSON)},
		},
		"temperature": 0.3,
		"max_tokens":  300,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", nvidiaBaseURL, bytes.NewReader(bodyBytes))
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NVIDIA NIM error: HTTP %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("NVIDIA NIM: empty response")
	}

	content := result.Choices[0].Message.Content
	if cacheKey != "" {
		global.Set(cacheKey, content)
	}
	return content, nil
}
