package repository

import (
	"context"
	"errors"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"

	"ukapp/gen/db/ukapp/public/model"
	. "ukapp/gen/db/ukapp/public/table"
)

// CreateEvent inserts a planned-works event and returns the stored row.
func (s *Store) CreateEvent(ctx context.Context, e model.Event) (model.Event, error) {
	var out model.Event
	err := Event.INSERT(Event.HouseID, Event.UkDispatcherID, Event.Entrance, Event.Riser, Event.Reason, Event.Responsible, Event.ScheduledFrom, Event.ScheduledTo).
		VALUES(e.HouseID, e.UkDispatcherID, e.Entrance, e.Riser, e.Reason, e.Responsible, e.ScheduledFrom, e.ScheduledTo).
		RETURNING(Event.AllColumns).
		QueryContext(ctx, s.db, &out)
	return out, err
}

// GetEvent returns ErrNotFound for a missing event.
func (s *Store) GetEvent(ctx context.Context, id int64) (model.Event, error) {
	var e model.Event
	err := SELECT(Event.AllColumns).FROM(Event).WHERE(Event.ID.EQ(Int64(id))).QueryContext(ctx, s.db, &e)
	if errors.Is(err, qrm.ErrNoRows) {
		return model.Event{}, ErrNotFound
	}
	return e, err
}

// ListEvents returns the house's events. activeOnly: planned/in_progress only, soonest first
// (banner); otherwise everything, latest scheduledFrom first (history).
func (s *Store) ListEvents(ctx context.Context, houseID int64, activeOnly bool) ([]model.Event, error) {
	where := Event.HouseID.EQ(Int64(houseID))
	order := Event.ScheduledFrom.DESC()
	if activeOnly {
		where = where.AND(Event.Status.IN(String("planned"), String("in_progress")))
		order = Event.ScheduledFrom.ASC()
	}
	var out []model.Event
	err := SELECT(Event.AllColumns).FROM(Event).WHERE(where).ORDER_BY(order).QueryContext(ctx, s.db, &out)
	return out, err
}
