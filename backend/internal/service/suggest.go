package service

import (
	"context"
	"strings"

	"ukapp/internal/dadata"
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
	return toAddressSuggestions(sugs), nil
}

// GeolocateAddresses возвращает ближайшие адреса с домом по координатам (reverse geocoding).
func (s *Service) GeolocateAddresses(ctx context.Context, lat, lon float64, count int) ([]AddressSuggestion, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, ErrInvalidInput
	}
	sugs, err := s.dd.Geolocate(ctx, lat, lon, clampSuggestCount(count))
	if err != nil {
		return nil, err
	}
	return toAddressSuggestions(sugs), nil
}

func toAddressSuggestions(sugs []dadata.Suggestion) []AddressSuggestion {
	out := make([]AddressSuggestion, 0, len(sugs))
	for _, sug := range sugs {
		out = append(out, AddressSuggestion{Value: sug.Value, HouseFiasID: sug.HouseFiasID})
	}
	return out
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
