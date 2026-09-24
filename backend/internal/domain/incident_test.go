package domain

import (
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

func TestFindDuplicate_NoRiserVsRiser(t *testing.T) {
	now := time.Now()
	open := []OpenIncident{{ID: 8, HouseID: 1, Category: CategoryElevator, Riser: nil, Status: IncidentAccepted, CreatedAt: now}}
	if got := FindDuplicate(1, CategoryElevator, ptr("1"), now, open); got != 0 {
		t.Fatalf("nil riser must not match a set one, got %d", got)
	}
	if got := FindDuplicate(1, CategoryElevator, nil, now, open); got != 8 {
		t.Fatalf("nil riser must match nil riser, got %d", got)
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
