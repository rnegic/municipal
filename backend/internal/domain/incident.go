package domain

import (
	"slices"
	"strings"
	"time"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
)

func (s Severity) Valid() bool {
	switch s {
	case SeverityCritical, SeverityWarning:
		return true
	}
	return false
}

type IncidentStatus string

const (
	IncidentAccepted   IncidentStatus = "accepted"
	IncidentInProgress IncidentStatus = "in_progress"
	IncidentVerifying  IncidentStatus = "verifying"
	IncidentDone       IncidentStatus = "done"
	IncidentFalseAlarm IncidentStatus = "false_alarm"
)

const (
	SuspiciousAfterFalseAlarms = 3
	MaxReportsPerHour          = 10
)

func (s IncidentStatus) Closed() bool {
	return s == IncidentDone || s == IncidentFalseAlarm
}

var sla = map[Severity]time.Duration{SeverityCritical: 4 * time.Hour, SeverityWarning: 24 * time.Hour}

func SLA(sev Severity) time.Duration { return sla[sev] }

const DedupWindow = 2 * time.Hour

const (
	DedupVersion  = "exact-category-floorzone-v2"
	ReportCreated = "created"
	ReportJoined  = "joined"
)

type OpenIncident struct {
	ID          int64
	HouseID     int64
	Category    Category
	Title       string
	Description string
	Entrance    *string
	Riser       *string
	Severity    Severity
	Status      IncidentStatus
	CreatedAt   time.Time
	Subscribed  bool
}

func FindDuplicate(houseID int64, category Category, riser *string, now time.Time, open []OpenIncident) int64 {
	for _, inc := range open {
		if inc.HouseID != houseID || inc.Category == "" || inc.Category != category || !sameRiser(inc.Riser, riser) {
			continue
		}
		if inc.Status != IncidentAccepted && inc.Status != IncidentInProgress {
			continue
		}
		if now.Sub(inc.CreatedAt) <= DedupWindow {
			return inc.ID
		}
	}
	return 0
}

var dispatcherFrom = map[IncidentStatus]IncidentStatus{
	IncidentInProgress: IncidentAccepted,
	IncidentVerifying:  IncidentInProgress,
	IncidentDone:       IncidentVerifying,
}

func DispatcherTransition(to IncidentStatus) (from IncidentStatus, ok bool) {
	from, ok = dispatcherFrom[to]
	return from, ok
}

var openStatuses = map[IncidentStatus]bool{
	IncidentAccepted:   true,
	IncidentInProgress: true,
	IncidentVerifying:  true,
}

func CanDispatcherMove(from, to IncidentStatus) bool {
	if from == to {
		return false
	}
	if openStatuses[from] && openStatuses[to] {
		return true
	}
	if from == IncidentVerifying && to == IncidentDone {
		return true
	}
	return from == IncidentDone && to == IncidentVerifying
}

func sameRiser(a, b *string) bool {
	x, y := normRiser(a), normRiser(b)
	return x == "" || y == "" || strings.EqualFold(x, y)
}

func normRiser(s *string) string {
	if s == nil {
		return ""
	}
	return strings.Join(strings.Fields(*s), " ")
}

const (
	DedupCandidateWindow = 7 * 24 * time.Hour
	MaxDedupCandidates   = 10
)

type DedupMatch struct {
	IncidentID   int64
	P            float64
	ModelVersion string
}

func DedupCandidates(open []OpenIncident, now time.Time) []OpenIncident {
	var out []OpenIncident
	for _, inc := range open {
		if (inc.Status == IncidentAccepted || inc.Status == IncidentInProgress) && now.Sub(inc.CreatedAt) <= DedupCandidateWindow {
			out = append(out, inc)
		}
	}
	slices.SortStableFunc(out, func(a, b OpenIncident) int { return b.CreatedAt.Compare(a.CreatedAt) })
	if len(out) > MaxDedupCandidates {
		out = out[:MaxDedupCandidates]
	}
	return out
}
