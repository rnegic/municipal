// Package domain — чистая доменная логика: дедуп, порог подтверждения «починили».
// Без I/O: принимает данные, возвращает решение.
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

// SLA per severity: dueAt = createdAt + SLA. ponytail: стартовые значения, согласовать с продуктом.
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

// FindDuplicate returns the id of an open incident a new report should join, or 0:
// same house, same title and riser, still open (accepted/in_progress), opened within DedupWindow.
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

// dispatcherFrom is the required current status for a dispatcher-requested transition to `to`.
var dispatcherFrom = map[IncidentStatus]IncidentStatus{
	IncidentInProgress: IncidentAccepted,
	IncidentVerifying:  IncidentInProgress,
	IncidentDone:       IncidentVerifying,
}

// DispatcherTransition returns the required current status for a dispatcher setting the
// incident to `to`, and false if `to` isn't a valid dispatcher-driven target.
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
