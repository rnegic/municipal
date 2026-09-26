package domain

import (
	"slices"
	"testing"
	"time"
)

func ptr(s string) *string { return &s }

func TestFindDuplicate(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	open := []OpenIncident{
		{ID: 1, HouseID: 10, Category: CategoryWaterHeat, Riser: ptr("1"), Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour)},
		{ID: 2, HouseID: 10, Category: CategoryElevator, Riser: ptr("1"), Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour)},
		{ID: 3, HouseID: 10, Category: CategoryWaterHeat, Riser: ptr("1"), Status: IncidentDone, CreatedAt: now.Add(-time.Hour)},
		{ID: 4, HouseID: 11, Category: CategoryWaterHeat, Riser: ptr("1"), Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour)},
		{ID: 5, HouseID: 12, Category: CategoryWaterHeat, Riser: ptr("1"), Status: IncidentInProgress, CreatedAt: now.Add(-3 * time.Hour)},
		{ID: 6, HouseID: 10, Category: CategoryWaterHeat, Riser: ptr("2"), Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour)},
	}
	cases := []struct {
		name  string
		house int64
		cat   Category
		riser *string
		want  int64
	}{
		{"same house+category+riser+open+fresh", 10, CategoryWaterHeat, ptr("1"), 1},
		{"different category", 10, CategoryCleaningYard, ptr("1"), 0},
		{"different house", 13, CategoryWaterHeat, ptr("1"), 0},
		{"different riser", 10, CategoryWaterHeat, ptr("3"), 0},
		{"done ignored", 10, CategoryWaterHeat, ptr("1"), 1},
		{"too old", 12, CategoryWaterHeat, ptr("1"), 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FindDuplicate(c.house, c.cat, c.riser, now, open); got != c.want {
				t.Fatalf("got %d want %d", got, c.want)
			}
		})
	}
}

func TestFindDuplicate_ExactlyTwoHoursIsStillDuplicate(t *testing.T) {
	now := time.Now()
	open := []OpenIncident{{ID: 7, HouseID: 1, Category: CategoryElevator, Riser: nil, Status: IncidentAccepted, CreatedAt: now.Add(-DedupWindow)}}
	if got := FindDuplicate(1, CategoryElevator, nil, now, open); got != 7 {
		t.Fatalf("got %d want 7", got)
	}
}

func TestFindDuplicate_RiserFuzzy(t *testing.T) {
	now := time.Now()
	open := []OpenIncident{{ID: 8, HouseID: 1, Category: CategoryElevator, Riser: ptr("  3 Этаж "), Status: IncidentAccepted, CreatedAt: now}}
	for _, r := range []*string{nil, ptr(""), ptr("3  этаж")} {
		if got := FindDuplicate(1, CategoryElevator, r, now, open); got != 8 {
			t.Fatalf("riser %v must match, got %d", r, got)
		}
	}
	if got := FindDuplicate(1, CategoryElevator, ptr("5 этаж"), now, open); got != 0 {
		t.Fatalf("different riser must not match, got %d", got)
	}
}

func TestSLA(t *testing.T) {
	if SLA(SeverityCritical) != 4*time.Hour || SLA(SeverityWarning) != 24*time.Hour {
		t.Error("SLA")
	}
}

func TestFindDuplicate_LegacyWithoutCategoryNeverMatches(t *testing.T) {
	now := time.Now()
	open := []OpenIncident{{ID: 9, HouseID: 1, Riser: nil, Status: IncidentAccepted, CreatedAt: now}}
	if got := FindDuplicate(1, "", nil, now, open); got != 0 {
		t.Fatalf("legacy incident without category matched: %d", got)
	}
}

func TestDedupCandidates(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	open := []OpenIncident{
		{ID: 1, Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour)},
		{ID: 2, Status: IncidentInProgress, CreatedAt: now.Add(-7 * 24 * time.Hour)},
		{ID: 3, Status: IncidentAccepted, CreatedAt: now.Add(-7*24*time.Hour - time.Second)},
		{ID: 4, Status: IncidentVerifying, CreatedAt: now},
		{ID: 5, Status: IncidentAccepted, Category: CategoryElevator, CreatedAt: now.Add(-time.Minute)},
	}
	got := DedupCandidates(open, now)
	var ids []int64
	for _, c := range got {
		ids = append(ids, c.ID)
	}
	if !slices.Equal(ids, []int64{5, 1, 2}) {
		t.Fatalf("got %v", ids)
	}
}

func TestDedupCandidatesLimit(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	var open []OpenIncident
	for i := range 25 {
		open = append(open, OpenIncident{ID: int64(i + 1), Status: IncidentAccepted, CreatedAt: now.Add(-time.Duration(i) * time.Minute)})
	}
	got := DedupCandidates(open, now)
	if len(got) != MaxDedupCandidates || got[0].ID != 1 || got[9].ID != 10 {
		t.Fatalf("len=%d first=%d last=%d", len(got), got[0].ID, got[len(got)-1].ID)
	}
}

func TestDedupCandidates_IncludesSubscribed(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	open := []OpenIncident{
		{ID: 1, Status: IncidentAccepted, CreatedAt: now.Add(-time.Hour), Subscribed: true},
		{ID: 2, Status: IncidentAccepted, CreatedAt: now.Add(-time.Minute)},
	}
	if got := DedupCandidates(open, now); len(got) != 2 {
		t.Fatalf("own incidents must reach the model: %+v", got)
	}
}

func TestCanDispatcherMove_PendingOnlyToAccepted(t *testing.T) {
	for _, to := range []IncidentStatus{IncidentInProgress, IncidentVerifying, IncidentDone, IncidentFalseAlarm} {
		if CanDispatcherMove(IncidentPending, to) {
			t.Fatalf("pending → %s must be forbidden", to)
		}
		if CanDispatcherMove(to, IncidentPending) {
			t.Fatalf("%s → pending must be forbidden", to)
		}
	}
	if !CanDispatcherMove(IncidentPending, IncidentAccepted) {
		t.Fatal("pending → accepted must be allowed")
	}
}
