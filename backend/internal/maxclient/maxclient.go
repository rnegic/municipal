package maxclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

func NewClient(token string) *Client {
	return &Client{Token: token, BaseURL: "https://platform-api2.max.ru", HTTP: &http.Client{Timeout: 10 * time.Second}}
}

func (c *Client) SendMessage(ctx context.Context, userID int64, text string) error {
	raw, _ := json.Marshal(map[string]any{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/messages?user_id="+strconv.FormatInt(userID, 10), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("max: status %d: %s", resp.StatusCode, b)
	}
	return nil
}

type Update struct {
	UpdateType string `json:"update_type"`
	User       struct {
		UserID    int64  `json:"user_id"`
		FirstName string `json:"first_name"`
		Name      string `json:"name"`
	} `json:"user"`
}

func (c *Client) Updates(ctx context.Context, marker *int64, types string) ([]Update, *int64, error) {
	url := c.BaseURL + "/updates?timeout=5&types=" + types
	if marker != nil {
		url += "&marker=" + strconv.FormatInt(*marker, 10)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, marker, err
	}
	req.Header.Set("Authorization", c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, marker, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, marker, fmt.Errorf("max: updates status %d: %s", resp.StatusCode, b)
	}
	var out struct {
		Updates []Update `json:"updates"`
		Marker  *int64   `json:"marker"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, marker, err
	}
	if out.Marker == nil {
		out.Marker = marker
	}
	return out.Updates, out.Marker, nil
}
