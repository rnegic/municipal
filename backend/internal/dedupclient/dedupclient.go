package dedupclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"ukapp/internal/domain"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type report struct {
	ID          int64   `json:"id,omitempty"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Entrance    *string `json:"entrance"`
	Riser       *string `json:"riser"`
	Severity    string  `json:"severity"`
	CreatedAt   string  `json:"created_at,omitempty"`
}

func toReport(o domain.OpenIncident) report {
	r := report{ID: o.ID, Title: o.Title, Description: o.Description, Entrance: o.Entrance, Riser: o.Riser, Severity: string(o.Severity)}
	if !o.CreatedAt.IsZero() {
		r.CreatedAt = o.CreatedAt.UTC().Format(time.RFC3339)
	}
	return r
}

func (c *Client) Match(ctx context.Context, req domain.OpenIncident, cands []domain.OpenIncident) (domain.DedupMatch, error) {
	rs := make([]report, len(cands))
	for i, o := range cands {
		rs[i] = toReport(o)
	}
	body, err := json.Marshal(map[string]any{"request": toReport(req), "candidates": rs, "top_k": len(cands)})
	if err != nil {
		return domain.DedupMatch{}, err
	}
	hr, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/match", bytes.NewReader(body))
	if err != nil {
		return domain.DedupMatch{}, err
	}
	hr.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(hr)
	if err != nil {
		return domain.DedupMatch{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return domain.DedupMatch{}, fmt.Errorf("dedup: status %d", resp.StatusCode)
	}
	var out struct {
		Match        *int64  `json:"match"`
		P            float64 `json:"duplicate_probability"`
		ModelVersion string  `json:"model_version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return domain.DedupMatch{}, fmt.Errorf("dedup: %w", err)
	}
	if out.ModelVersion == "" {
		return domain.DedupMatch{}, fmt.Errorf("dedup: empty model_version")
	}
	m := domain.DedupMatch{P: out.P, ModelVersion: out.ModelVersion}
	if out.Match != nil {
		if !slices.ContainsFunc(cands, func(o domain.OpenIncident) bool { return o.ID == *out.Match }) {
			return domain.DedupMatch{}, fmt.Errorf("dedup: match %d is not a candidate", *out.Match)
		}
		m.IncidentID = *out.Match
	}
	return m, nil
}
