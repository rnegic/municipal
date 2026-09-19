package repository

import (
	"context"

	. "github.com/go-jet/jet/v2/postgres"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

// UnregisteredRow = incident + house.external_id; alias must match the field name.
type UnregisteredRow struct {
	model.Incident
	HouseExternalID *string
}

// UnregisteredIncidents returns incidents not yet known to the UK system, oldest first.
func (s *Store) UnregisteredIncidents(ctx context.Context, limit int64) ([]UnregisteredRow, error) {
	var rows []UnregisteredRow
	err := SELECT(Incident.AllColumns, House.ExternalID.AS("unregistered_row.house_external_id")).
		FROM(Incident.INNER_JOIN(House, House.ID.EQ(Incident.HouseID))).
		WHERE(Incident.ExternalID.IS_NULL()).
		ORDER_BY(Incident.ID).LIMIT(limit).
		QueryContext(ctx, s.db, &rows)
	return rows, err
}

func (s *Store) SetIncidentExternalID(ctx context.Context, id int64, externalID string) error {
	_, err := Incident.UPDATE(Incident.ExternalID).SET(externalID).
		WHERE(Incident.ID.EQ(Int64(id))).ExecContext(ctx, s.db)
	return err
}
