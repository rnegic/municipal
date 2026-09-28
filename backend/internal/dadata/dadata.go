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
	Token        string
	BaseURL      string
	GeolocateURL string
	HTTP         *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		Token:        token,
		BaseURL:      "https://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address",
		GeolocateURL: "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address",
		HTTP:         &http.Client{Timeout: 5 * time.Second},
	}
}

type Suggestion struct {
	Value        string
	HouseFiasID  string
	RegionFiasID string
	StreetFiasID string
	House        string
}

func (c *Client) Suggest(ctx context.Context, q string, count int) ([]Suggestion, error) {
	return c.post(ctx, c.BaseURL, map[string]any{"query": q, "count": normalizeCount(count)})
}

func (c *Client) Geolocate(ctx context.Context, lat, lon float64, count int) ([]Suggestion, error) {
	return c.post(ctx, c.GeolocateURL, map[string]any{"lat": lat, "lon": lon, "count": normalizeCount(count)})
}

func (c *Client) post(ctx context.Context, url string, payload map[string]any) ([]Suggestion, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
				HouseFiasID  *string `json:"house_fias_id"`
				RegionFiasID *string `json:"region_fias_id"`
				StreetFiasID *string `json:"street_fias_id"`
				House        *string `json:"house"`
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
		res = append(res, Suggestion{
			Value: s.Value, HouseFiasID: *s.Data.HouseFiasID,
			RegionFiasID: deref(s.Data.RegionFiasID), StreetFiasID: deref(s.Data.StreetFiasID), House: deref(s.Data.House),
		})
	}
	return res, nil
}

func normalizeCount(count int) int {
	if count < 1 {
		return defaultSuggestCount
	}
	return count
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
