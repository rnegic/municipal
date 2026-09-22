package domain

import "time"

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
)

var sla = map[Severity]time.Duration{SeverityCritical: 4 * time.Hour, SeverityWarning: 24 * time.Hour}

func SLA(sev Severity) time.Duration { return sla[sev] }

const DedupWindow = 2 * time.Hour

type OpenIncident struct {
	ID        int64
	HouseID   int64
	Title     string
	Riser     *string
	Status    IncidentStatus
	CreatedAt time.Time
}

func FindDuplicate(houseID int64, title string, riser *string, now time.Time, open []OpenIncident) int64 {
	for _, inc := range open {
		if inc.HouseID != houseID || inc.Title != title || !sameRiser(inc.Riser, riser) {
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

func sameRiser(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
