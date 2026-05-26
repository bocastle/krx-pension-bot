package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultTelegramBaseURL = "https://api.telegram.org"

type Client struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewClient(token, baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = defaultTelegramBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		token:   token,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    httpClient,
	}
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	body := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage failed: status %d", resp.StatusCode)
	}
	return nil
}
