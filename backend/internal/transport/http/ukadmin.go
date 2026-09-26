package http

import (
	"context"
	"errors"
	"strconv"
	"time"

	oapi "ukapp/gen/api"
	"ukapp/internal/domain"
	"ukapp/internal/service"

	"ukapp/gen/db/ukapp/public/model"
)

func formatUkID(id int64) string    { return "uk_" + strconv.FormatInt(id, 10) }
func formatEventID(id int64) string { return "ev_" + strconv.FormatInt(id, 10) }

const ukLoginDenied = "неверный ИНН или пароль, либо у организации нет доступа к АРМ"

func (s *server) EsiaMockLogin(ctx context.Context, req oapi.EsiaMockLoginRequestObject) (oapi.EsiaMockLoginResponseObject, error) {
	sess, err := s.svc.EsiaMockLogin(ctx, req.Body.Inn, req.Body.Password, maxUserFromCtx(ctx))
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return oapi.EsiaMockLogin400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "ИНН организации — 10 цифр с верным контрольным разрядом, пароль обязателен"))}, nil
	case errors.Is(err, service.ErrForbidden):
		return oapi.EsiaMockLogin403JSONResponse(apiErr("forbidden", ukLoginDenied)), nil
	case errors.Is(err, service.ErrRateLimited):
		return oapi.EsiaMockLogin429JSONResponse(apiErr("rate_limited", "слишком много попыток входа, попробуйте через 15 минут")), nil
	case err != nil:
		return nil, err
	}
	position := ""
	if sess.User.Position != nil {
		position = *sess.User.Position
	}
	return oapi.EsiaMockLogin200JSONResponse{
		Token: sess.Token, ExpiresAt: sess.ExpiresAt, AuthMethod: oapi.EsiaMock,
		User: oapi.UkDispatcher{
			Id: formatUserID(sess.User.ID), FullName: sess.User.FullName, Position: position,
			Role: oapi.UkDispatcherRoleUkDispatcher,
		},
		Organization: oapi.UkOrganization{
			Id: formatUkID(sess.Org.ID), Inn: *sess.Org.Inn, Ogrn: sess.Org.Ogrn, Name: sess.Org.Name,
			LicenseNumber: sess.Org.LicenseNumber,
		},
	}, nil
}

func (s *server) UnbindHouse(ctx context.Context, _ oapi.UnbindHouseRequestObject) (oapi.UnbindHouseResponseObject, error) {
	if err := s.svc.UnbindHouse(ctx, userFromCtx(ctx).ID); err != nil {
		return nil, err
	}
	return oapi.UnbindHouse204Response{}, nil
}

func (s *server) ListUkQueue(ctx context.Context, req oapi.ListUkQueueRequestObject) (oapi.ListUkQueueResponseObject, error) {
	offset, limit := int64(0), int64(100)
	if req.Params.Offset != nil {
		offset = int64(*req.Params.Offset)
	}
	if req.Params.Limit != nil {
		limit = int64(*req.Params.Limit)
	}
	rows, total, err := s.svc.UkQueue(ctx, userFromCtx(ctx), offset, limit)
	if err != nil {
		return nil, err
	}
	items := make([]oapi.UkQueueItem, len(rows))
	for i, r := range rows {
		items[i] = oapi.UkQueueItem{
			Id: formatIncidentID(r.ID), HouseId: formatHouseID(r.HouseID), HouseAddress: r.HouseAddress,
			Title: r.Title, Description: r.Description, Severity: oapi.Severity(r.Severity), Status: publicStatus(r.Status),
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, AffectedCount: r.AffectedCount, ConfirmedCount: r.ConfirmedCount,
			ReporterName: r.ReporterName, Photos: toPhotos(r.PhotoIDs),
			Category: (*oapi.IncidentCategory)(r.Category), MergedCount: r.MergedCount,
		}
	}
	return oapi.ListUkQueue200JSONResponse{Items: items, Total: int(total), Offset: int(offset), Limit: int(limit)}, nil
}

func toEvent(e model.UkEvent, now time.Time) oapi.Event {
	status := oapi.EventStatusClosed
	switch {
	case now.Before(e.ScheduledFrom):
		status = oapi.EventStatusPlanned
	case now.Before(e.ScheduledTo):
		status = oapi.EventStatusInProgress
	}
	return oapi.Event{
		Id: formatEventID(e.ID), HouseId: formatHouseID(e.HouseID), Entrance: e.Entrance, Riser: e.Riser,
		Reason: e.Reason, Responsible: e.Responsible, ScheduledFrom: e.ScheduledFrom, ScheduledTo: e.ScheduledTo,
		Status: status, CreatedAt: e.CreatedAt,
	}
}

func (s *server) CreateUkEvent(ctx context.Context, req oapi.CreateUkEventRequestObject) (oapi.CreateUkEventResponseObject, error) {
	houseID, ok := parseID("h_", req.Body.HouseId)
	if !ok {
		return oapi.CreateUkEvent404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	b := req.Body
	e, err := s.svc.CreateEvent(ctx, userFromCtx(ctx), service.EventInput{
		HouseID: houseID, Reason: b.Reason, Responsible: b.Responsible, Entrance: b.Entrance, Riser: b.Riser,
		ScheduledFrom: b.ScheduledFrom, ScheduledTo: b.ScheduledTo,
	})
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return oapi.CreateUkEvent400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "нужны причина, ответственный и scheduledTo позже scheduledFrom"))}, nil
	case errors.Is(err, service.ErrNotFound):
		return oapi.CreateUkEvent404JSONResponse(apiErr("not_found", "дом не найден")), nil
	case err != nil:
		return nil, err
	}
	return oapi.CreateUkEvent201JSONResponse(toEvent(e, time.Now())), nil
}

func (s *server) ListEvents(ctx context.Context, req oapi.ListEventsRequestObject) (oapi.ListEventsResponseObject, error) {
	houseID, ok := parseID("h_", req.Params.HouseId)
	if !ok {
		return oapi.ListEvents404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	events, err := s.svc.ListEvents(ctx, userFromCtx(ctx), houseID)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.ListEvents404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]oapi.Event, len(events))
	for i, e := range events {
		items[i] = toEvent(e, now)
	}
	return oapi.ListEvents200JSONResponse{Items: items}, nil
}

func (s *server) MergeIncidents(ctx context.Context, req oapi.MergeIncidentsRequestObject) (oapi.MergeIncidentsResponseObject, error) {
	notFound := oapi.MergeIncidents404JSONResponse(apiErr("not_found", "заявка не найдена"))
	target, ok := parseID("inc_", req.Body.TargetIncidentId)
	if !ok {
		return notFound, nil
	}
	sources := make([]int64, len(req.Body.SourceIncidentIds))
	for i, raw := range req.Body.SourceIncidentIds {
		if sources[i], ok = parseID("inc_", raw); !ok {
			return notFound, nil
		}
	}
	res, err := s.svc.MergeIncidents(ctx, userFromCtx(ctx), target, sources)
	switch {
	case errors.Is(err, domain.ErrMergeInvalid):
		return oapi.MergeIncidents400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "нужна хотя бы одна заявка-дубликат, отличная от главной"))}, nil
	case errors.Is(err, domain.ErrMergeNotFound):
		return notFound, nil
	case errors.Is(err, domain.ErrMergeForeign):
		return oapi.MergeIncidents403JSONResponse(apiErr("forbidden", "заявка относится к дому другой УК")), nil
	case errors.Is(err, domain.ErrMergeRuleFails):
		return oapi.MergeIncidents422JSONResponse(apiErr("business_rule_failed", "склеивать можно только незакрытые заявки одного дома")), nil
	case err != nil:
		return nil, err
	}
	merged := make([]string, 0, len(sources))
	for _, id := range sources {
		merged = append(merged, formatIncidentID(id))
	}
	return oapi.MergeIncidents200JSONResponse{
		TargetIncidentId: formatIncidentID(target), MergedIncidentIds: merged, AffectedCount: res.AffectedCount,
	}, nil
}
