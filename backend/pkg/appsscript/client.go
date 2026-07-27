package appsscript

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

type appsScriptResponse struct {
	Success bool                     `json:"success"`
	Data    []map[string]interface{} `json:"data"`
}

// TODO(D2): accept context.Context, use http.NewRequestWithContext
func (c *Client) GetSheet(sheet string) ([]map[string]interface{}, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid apps script URL: %w", err)
	}
	q := u.Query()
	q.Set("action", sheet)
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("get sheet %s: %w", sheet, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read sheet %s: %w", sheet, err)
	}

	var sr appsScriptResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("decode sheet %s: %w", sheet, err)
	}

	return sr.Data, nil
}

func (c *Client) SetValueByKey(key, value string) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid apps script URL: %w", err)
	}
	q := u.Query()
	q.Set("action", "setValueByKey")
	q.Set("sheet", "config")
	q.Set("key", key)
	q.Set("value", value)
	u.RawQuery = q.Encode()

	resp, err := c.httpClient.Post(u.String(), "application/x-www-form-urlencoded", nil)
	if err != nil {
		return fmt.Errorf("set value %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("set value %s: status %d: %s", key, resp.StatusCode, string(b))
	}

	return nil
}
