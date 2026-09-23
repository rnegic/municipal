package dadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultSuggestCount = 5

type Client struct {
	Token   string
	BaseURL string
	HTTP    *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token:   token,
		BaseURL: "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address",
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

type Suggestion struct {
	Value       string
	HouseFiasID string
}

func (c *Client) Suggest(ctx context.Context, q string, count int) ([]Suggestion, error) {
	if count < 1 {
		count = defaultSuggestCount
	}
	body, _ := json.Marshal(map[string]any{"query": q, "count": count})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dadata: status %d", resp.StatusCode)
	}
	var out struct {
		Suggestions []struct {
			Value string `json:"value"`
			Data  struct {
				HouseFiasID *string `json:"house_fias_id"`
			} `json:"data"`
		} `json:"suggestions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	res := []Suggestion{}
	for _, s := range out.Suggestions {
		if s.Data.HouseFiasID == nil {
			continue
		}
		res = append(res, Suggestion{Value: s.Value, HouseFiasID: *s.Data.HouseFiasID})
	}
	return res, nil
}
