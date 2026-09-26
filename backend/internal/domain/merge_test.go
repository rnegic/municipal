package domain

import (
	"errors"
	"slices"
	"testing"
)

func TestMergeIDs(t *testing.T) {
	if _, err := MergeIDs(1, nil); !errors.Is(err, ErrMergeInvalid) {
		t.Fatalf("empty: %v", err)
	}
	if _, err := MergeIDs(1, []int64{2, 1}); !errors.Is(err, ErrMergeInvalid) {
		t.Fatalf("contains target: %v", err)
	}
	if got, _ := MergeIDs(1, []int64{3, 2, 3}); !slices.Equal(got, []int64{3, 2}) {
		t.Fatalf("dedup: %v", got)
	}
}

func TestPlanMerge(t *testing.T) {
	one := int64(1)
	nine := int64(9)
	open := func(id int64) MergeCandidate {
		return MergeCandidate{ID: id, HouseID: 10, UkID: 7, Status: IncidentAccepted}
	}
	with := func(c MergeCandidate, f func(*MergeCandidate)) MergeCandidate { f(&c); return c }

	cases := []struct {
		name  string
		found []MergeCandidate
		want  []int64
		err   error
	}{
		{"ok", []MergeCandidate{open(1), open(2), open(3)}, []int64{2, 3}, nil},
		{"missing source", []MergeCandidate{open(1), open(2)}, nil, ErrMergeNotFound},
		{"missing target", []MergeCandidate{open(2), open(3)}, nil, ErrMergeNotFound},
		{"foreign uk", []MergeCandidate{open(1), open(2), with(open(3), func(c *MergeCandidate) { c.UkID = 8 })}, nil, ErrMergeForeign},
		{"other house", []MergeCandidate{open(1), open(2), with(open(3), func(c *MergeCandidate) { c.HouseID = 11 })}, nil, ErrMergeRuleFails},
		{"target done", []MergeCandidate{with(open(1), func(c *MergeCandidate) { c.Status = IncidentDone }), open(2), open(3)}, nil, ErrMergeRuleFails},
		{"source done", []MergeCandidate{open(1), open(2), with(open(3), func(c *MergeCandidate) { c.Status = IncidentDone })}, nil, ErrMergeRuleFails},
		{"already merged here is skipped", []MergeCandidate{open(1), open(2), with(open(3), func(c *MergeCandidate) {
			c.Status, c.MergedIntoID = IncidentDone, &one
		})}, []int64{2}, nil},
		{"merged elsewhere", []MergeCandidate{open(1), open(2), with(open(3), func(c *MergeCandidate) {
			c.Status, c.MergedIntoID = IncidentDone, &nine
		})}, nil, ErrMergeRuleFails},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := PlanMerge(7, 1, []int64{2, 3}, c.found)
			if !errors.Is(err, c.err) || !slices.Equal(got, c.want) {
				t.Fatalf("got %v, %v; want %v, %v", got, err, c.want, c.err)
			}
		})
	}
}

func TestShortName(t *testing.T) {
	for in, want := range map[string]string{"Анна Михайлова": "Анна М.", "Анна": "Анна", " анна  петрова ": "анна П."} {
		if got := ShortName(in); got == nil || *got != want {
			t.Errorf("%q: got %v want %q", in, got, want)
		}
	}
	if ShortName("  ") != nil {
		t.Error("blank name must be nil")
	}
}
