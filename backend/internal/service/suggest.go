package service

import (
	"context"
	"strings"
)

const (
	DefaultSuggestCount = 5
	MaxSuggestCount     = 10
)

type AddressSuggestion struct {
	Value       string
	HouseFiasID string
}

func (s *Service) SuggestAddresses(ctx context.Context, query string, count int) ([]AddressSuggestion, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, ErrInvalidInput
	}
	sugs, err := s.dd.Suggest(ctx, q, clampSuggestCount(count))
	if err != nil {
		return nil, err
	}
	out := make([]AddressSuggestion, 0, len(sugs))
	for _, sug := range sugs {
		out = append(out, AddressSuggestion{Value: sug.Value, HouseFiasID: sug.HouseFiasID})
	}
	return out, nil
}

func clampSuggestCount(count int) int {
	switch {
	case count < 1:
		return DefaultSuggestCount
	case count > MaxSuggestCount:
		return MaxSuggestCount
	default:
		return count
	}
}
