package repository

import (
	"context"

	"ukapp/internal/domain"

	. "ukapp/gen/db/ukapp/public/table"
)

type ReportInput struct {
	IncidentID   int64
	ReporterID   int64
	HouseID      int64
	Title        string
	Description  string
	Severity     domain.Severity
	Entrance     *string
	Riser        *string
	Outcome      string
	DedupVersion string
}

func (s *Store) AddReport(ctx context.Context, r ReportInput) error {
	_, err := IncidentReport.INSERT(
		IncidentReport.IncidentID, IncidentReport.ReporterID, IncidentReport.HouseID,
		IncidentReport.Title, IncidentReport.Description, IncidentReport.Severity,
		IncidentReport.Entrance, IncidentReport.Riser, IncidentReport.Outcome, IncidentReport.DedupVersion,
	).
		VALUES(r.IncidentID, r.ReporterID, r.HouseID, r.Title, r.Description, string(r.Severity), r.Entrance, r.Riser, r.Outcome, r.DedupVersion).
		ON_CONFLICT().DO_NOTHING().
		ExecContext(ctx, s.db)
	return err
}
