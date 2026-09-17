package http

import (
	"context"
	"errors"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"
)

func toIncident(r service.IncidentRow) oapi.Incident {
	return oapi.Incident{
		Id: formatIncidentID(r.ID), HouseId: formatHouseID(r.HouseID),
		Title: r.Title, Description: r.Description, Severity: oapi.Severity(r.Severity),
		Status: oapi.IncidentStatus(r.Status), AffectedCount: r.AffectedCount,
		CreatedAt: r.CreatedAt, JoinedByMe: r.JoinedByMe, ConfirmedByMe: r.ConfirmedByMe,
	}
}

func (s *server) CreateIncident(ctx context.Context, req oapi.CreateIncidentRequestObject) (oapi.CreateIncidentResponseObject, error) {
	u := userFromCtx(ctx)
	if u.HouseID == nil {
		return oapi.CreateIncident404JSONResponse(apiErr("not_found", "дом не привязан")), nil
	}
	row, created, err := s.svc.CreateIncident(ctx, *u.HouseID, u.ID, req.Body.Title, req.Body.Description, string(req.Body.Severity), req.Body.Entrance, req.Body.Riser)
	if errors.Is(err, service.ErrInvalidInput) {
		return oapi.CreateIncident400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "valid title, description and severity required"))}, nil
	}
	if err != nil {
		return nil, err
	}
	if created {
		return oapi.CreateIncident201JSONResponse(toIncident(row)), nil
	}
	return oapi.CreateIncident200JSONResponse(toIncident(row)), nil
}

func (s *server) GetIncident(ctx context.Context, req oapi.GetIncidentRequestObject) (oapi.GetIncidentResponseObject, error) {
	id, ok := parseID("inc_", req.Id)
	if !ok {
		return oapi.GetIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	row, err := s.svc.GetIncident(ctx, id, userFromCtx(ctx).ID)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.GetIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.GetIncident200JSONResponse(toIncident(row)), nil
}

func (s *server) JoinIncident(ctx context.Context, req oapi.JoinIncidentRequestObject) (oapi.JoinIncidentResponseObject, error) {
	id, ok := parseID("inc_", req.Id)
	if !ok {
		return oapi.JoinIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	affected, joined, err := s.svc.JoinIncident(ctx, id, userFromCtx(ctx).ID)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.JoinIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.JoinIncident200JSONResponse{IncidentId: req.Id, AffectedCount: affected, Joined: joined}, nil
}

func (s *server) ConfirmIncident(ctx context.Context, req oapi.ConfirmIncidentRequestObject) (oapi.ConfirmIncidentResponseObject, error) {
	id, ok := parseID("inc_", req.Id)
	if !ok {
		return oapi.ConfirmIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	status, confirmedAt, err := s.svc.ConfirmIncident(ctx, id, userFromCtx(ctx).ID)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.ConfirmIncident404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	if errors.Is(err, service.ErrInvalidStatus) {
		return oapi.ConfirmIncident422JSONResponse(apiErr("business_rule_failed", "нельзя подтвердить, пока статус не verifying")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.ConfirmIncident200JSONResponse{IncidentId: req.Id, Status: oapi.IncidentStatus(status), ConfirmedAt: confirmedAt}, nil
}

func (s *server) ListHouseIncidents(ctx context.Context, req oapi.ListHouseIncidentsRequestObject) (oapi.ListHouseIncidentsResponseObject, error) {
	houseID, ok := parseID("h_", req.HouseId)
	u := userFromCtx(ctx)
	if !ok || u.HouseID == nil || *u.HouseID != houseID {
		return oapi.ListHouseIncidents404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	rows, err := s.svc.ListActiveIncidents(ctx, houseID, u.ID)
	if err != nil {
		return nil, err
	}
	items := make([]oapi.Incident, len(rows))
	for i, r := range rows {
		items[i] = toIncident(r)
	}
	return oapi.ListHouseIncidents200JSONResponse{Items: items}, nil
}

func (s *server) ListHouseRequests(ctx context.Context, req oapi.ListHouseRequestsRequestObject) (oapi.ListHouseRequestsResponseObject, error) {
	houseID, ok := parseID("h_", req.HouseId)
	u := userFromCtx(ctx)
	if !ok || u.HouseID == nil || *u.HouseID != houseID {
		return oapi.ListHouseRequests404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	offset, limit := int64(0), int64(20)
	if req.Params.Offset != nil {
		offset = int64(*req.Params.Offset)
	}
	if req.Params.Limit != nil {
		limit = int64(*req.Params.Limit)
	}
	rows, total, err := s.svc.ListRequests(ctx, houseID, u.ID, offset, limit)
	if err != nil {
		return nil, err
	}
	items := make([]oapi.ResidentRequest, len(rows))
	for i, r := range rows {
		items[i] = oapi.ResidentRequest{
			Id: formatRequestID(r.ID), Title: r.Title, Status: oapi.IncidentStatus(r.Status),
			CreatedAt: r.CreatedAt, DueAt: nil,
		}
	}
	return oapi.ListHouseRequests200JSONResponse{Items: items, Total: int(total), Offset: int(offset), Limit: int(limit)}, nil
}
