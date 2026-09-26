package domain

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const (
	ReportMergedManual = "merged_manual"
	MergeDedupVersion  = "manual-uk"
)

var (
	ErrMergeInvalid   = errors.New("merge: invalid request")
	ErrMergeNotFound  = errors.New("merge: incident not found")
	ErrMergeForeign   = errors.New("merge: incident of another uk")
	ErrMergeRuleFails = errors.New("merge: business rule failed")
)

type MergeCandidate struct {
	ID, HouseID, UkID int64
	Status            IncidentStatus
	MergedIntoID      *int64
}

func MergeIDs(target int64, sources []int64) ([]int64, error) {
	seen := map[int64]bool{}
	out := make([]int64, 0, len(sources))
	for _, id := range sources {
		if id == target {
			return nil, ErrMergeInvalid
		}
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	if len(out) == 0 {
		return nil, ErrMergeInvalid
	}
	return out, nil
}

func PlanMerge(ukID, target int64, sources []int64, found []MergeCandidate) (toMerge []int64, err error) {
	byID := make(map[int64]MergeCandidate, len(found))
	for _, c := range found {
		byID[c.ID] = c
	}
	t, ok := byID[target]
	if !ok {
		return nil, ErrMergeNotFound
	}
	for _, id := range sources {
		if _, ok := byID[id]; !ok {
			return nil, ErrMergeNotFound
		}
	}
	for _, c := range found {
		if c.UkID != ukID {
			return nil, ErrMergeForeign
		}
	}
	if t.Status.Closed() {
		return nil, ErrMergeRuleFails
	}
	for _, id := range sources {
		s := byID[id]
		if s.HouseID != t.HouseID {
			return nil, ErrMergeRuleFails
		}
		if s.MergedIntoID != nil && *s.MergedIntoID == target {
			continue
		}
		if s.MergedIntoID != nil || s.Status.Closed() {
			return nil, ErrMergeRuleFails
		}
		toMerge = append(toMerge, id)
	}
	return toMerge, nil
}

func ShortName(fullName string) *string {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return nil
	}
	name := parts[0]
	if len(parts) > 1 {
		r, _ := utf8.DecodeRuneInString(parts[len(parts)-1])
		name += " " + strings.ToUpper(string(r)) + "."
	}
	return &name
}
