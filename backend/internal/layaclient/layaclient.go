package layaclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"ukapp/internal/domain"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client { return &Client{baseURL: baseURL, http: &http.Client{}} }

func (c *Client) Classify(ctx context.Context, text string) (domain.Prediction, error) {
	body, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return domain.Prediction{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/classify", bytes.NewReader(body))
	if err != nil {
		return domain.Prediction{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return domain.Prediction{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.Prediction{}, fmt.Errorf("laya: status %d", resp.StatusCode)
	}
	var out struct {
		Category struct {
			Choice        string             `json:"choice"`
			Probabilities map[string]float64 `json:"probabilities"`
		} `json:"category"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.Prediction{}, fmt.Errorf("laya: %w", err)
	}
	p := domain.Prediction{Category: domain.Category(out.Category.Choice), P: out.Category.Probabilities[out.Category.Choice]}
	if !p.Category.Valid() {
		return domain.Prediction{}, fmt.Errorf("laya: unknown category %q", p.Category)
	}
	return p, nil
}
