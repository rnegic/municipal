package http

import (
	"context"
	"errors"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"

	"ukapp/gen/db/ukapp/public/model"
)

func toEvent(e model.Event) oapi.Event {
	return oapi.Event{
		Id: formatEventID(e.ID), HouseId: formatHouseID(e.HouseID), Entrance: e.Entrance, Riser: e.Riser,
		Reason: e.Reason, Responsible: e.Responsible, ScheduledFrom: e.ScheduledFrom, ScheduledTo: e.ScheduledTo,
		Status: oapi.EventStatus(e.Status), CreatedAt: e.CreatedAt, ResolvedAt: e.ResolvedAt,
	}
}

func (s *server) ListEvents(ctx context.Context, req oapi.ListEventsRequestObject) (oapi.ListEventsResponseObject, error) {
	houseID, ok := parseID("h_", req.Params.HouseId)
	if ok {
		var err error
		if ok, err = s.svc.CanAccessHouse(ctx, userFromCtx(ctx), houseID); err != nil {
			return nil, err
		}
	}
	if !ok {
		return oapi.ListEvents404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	rows, err := s.svc.ListEvents(ctx, houseID, req.Params.Scope != nil)
	if err != nil {
		return nil, err
	}
	items := make([]oapi.Event, len(rows))
	for i, e := range rows {
		items[i] = toEvent(e)
	}
	return oapi.ListEvents200JSONResponse{Items: items}, nil
}

func (s *server) GetEvent(ctx context.Context, req oapi.GetEventRequestObject) (oapi.GetEventResponseObject, error) {
	id, ok := parseID("evt_", req.Id)
	if !ok {
		return oapi.GetEvent404JSONResponse(apiErr("not_found", "событие не найдено")), nil
	}
	e, err := s.svc.GetEvent(ctx, userFromCtx(ctx), id)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.GetEvent404JSONResponse(apiErr("not_found", "событие не найдено")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.GetEvent200JSONResponse(toEvent(e)), nil
}

func (s *server) UkCreateEvent(ctx context.Context, req oapi.UkCreateEventRequestObject) (oapi.UkCreateEventResponseObject, error) {
	houseID, ok := parseID("h_", req.Body.HouseId)
	if !ok {
		return oapi.UkCreateEvent404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	b := req.Body
	e, err := s.svc.CreateEvent(ctx, userFromCtx(ctx), houseID, b.Entrance, b.Riser, b.Reason, b.Responsible, b.ScheduledFrom, b.ScheduledTo)
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return oapi.UkCreateEvent400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "reason, responsible required; scheduledTo must be after scheduledFrom"))}, nil
	case errors.Is(err, service.ErrNotFound):
		return oapi.UkCreateEvent404JSONResponse(apiErr("not_found", "дом не найден")), nil
	case err != nil:
		return nil, err
	}
	return oapi.UkCreateEvent201JSONResponse(toEvent(e)), nil
}
