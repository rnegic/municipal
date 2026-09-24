package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"

	oapi "ukapp/gen/api"
	"ukapp/internal/domain"
	"ukapp/internal/service"
)

func toIncident(r service.IncidentRow) oapi.Incident {
	return oapi.Incident{
		Id: formatIncidentID(r.ID), HouseId: formatHouseID(r.HouseID),
		Title: r.Title, Description: r.Description, Entrance: r.Entrance, Riser: r.Riser, Severity: oapi.Severity(r.Severity),
		Status: oapi.IncidentStatus(r.Status), AffectedCount: r.AffectedCount,
		CreatedAt: r.CreatedAt, DueAt: r.DueAt, JoinedByMe: r.JoinedByMe, ConfirmedByMe: r.ConfirmedByMe,
		Photos: toPhotos(r.PhotoIDs),
	}
}

func toPhoto(id int64) oapi.Photo {
	return oapi.Photo{Id: formatPhotoID(id), Url: "/api/photos/" + formatPhotoID(id)}
}

func toPhotos(ids []int64) []oapi.Photo {
	photos := make([]oapi.Photo, len(ids))
	for i, id := range ids {
		photos[i] = toPhoto(id)
	}
	return photos
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
	u := userFromCtx(ctx)
	err := s.svc.CanAccessIncident(ctx, u, id)
	var row service.IncidentRow
	if err == nil {
		row, err = s.svc.GetIncident(ctx, id, u.ID)
	}
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

func (s *server) SetIncidentStatus(ctx context.Context, req oapi.SetIncidentStatusRequestObject) (oapi.SetIncidentStatusResponseObject, error) {
	id, ok := parseID("inc_", req.Id)
	if !ok {
		return oapi.SetIncidentStatus404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	row, err := s.svc.SetIncidentStatus(ctx, userFromCtx(ctx), id, domain.IncidentStatus(req.Body.Status))
	if errors.Is(err, service.ErrNotFound) {
		return oapi.SetIncidentStatus404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	if errors.Is(err, service.ErrInvalidStatus) {
		return oapi.SetIncidentStatus422JSONResponse(apiErr("business_rule_failed", "недопустимый переход статуса")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.SetIncidentStatus200JSONResponse(toIncident(row)), nil
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
			Id: formatIncidentID(r.ID), Title: r.Title, Status: oapi.IncidentStatus(r.Status),
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, ConfirmedByMe: r.ConfirmedByMe,
		}
	}
	return oapi.ListHouseRequests200JSONResponse{Items: items, Total: int(total), Offset: int(offset), Limit: int(limit)}, nil
}

const maxPhotoBytes = 10 << 20

func readPhoto(body *multipart.Reader) (data []byte, contentType, problem string) {
	for {
		part, err := body.NextPart()
		if errors.Is(err, io.EOF) {
			return nil, "", "multipart field photo is required"
		}
		if err != nil {
			return nil, "", err.Error()
		}
		if part.FormName() != "photo" {
			continue
		}
		if data, err = io.ReadAll(io.LimitReader(part, maxPhotoBytes+1)); err != nil {
			return nil, "", err.Error()
		}
		break
	}
	if len(data) == 0 || len(data) > maxPhotoBytes {
		return nil, "", "photo must be 1 byte .. 10 MB"
	}
	if ct := http.DetectContentType(data); ct == "image/jpeg" || ct == "image/png" {
		return data, ct, ""
	}
	if !isHEIF(data) {
		return nil, "", "photo must be image/jpeg, image/png or image/heic"
	}
	jpg, err := heicToJPEG(data)
	if err != nil {
		return nil, "", "cannot decode heic"
	}
	return jpg, "image/jpeg", ""
}

func (s *server) UploadStagedPhoto(ctx context.Context, req oapi.UploadStagedPhotoRequestObject) (oapi.UploadStagedPhotoResponseObject, error) {
	data, ct, problem := readPhoto(req.Body)
	if problem != "" {
		return oapi.UploadStagedPhoto400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", problem))}, nil
	}
	id, err := s.svc.AddStagedPhoto(ctx, userFromCtx(ctx).ID, ct, data)
	if errors.Is(err, service.ErrRateLimited) {
		return oapi.UploadStagedPhoto429JSONResponse(apiErr("rate_limited", "не больше 20 фото в час")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.UploadStagedPhoto201JSONResponse(toPhoto(id)), nil
}

func (s *server) UploadIncidentPhoto(ctx context.Context, req oapi.UploadIncidentPhotoRequestObject) (oapi.UploadIncidentPhotoResponseObject, error) {
	id, ok := parseID("inc_", req.Id)
	if !ok {
		return oapi.UploadIncidentPhoto404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	}
	data, ct, problem := readPhoto(req.Body)
	if problem != "" {
		return oapi.UploadIncidentPhoto400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", problem))}, nil
	}
	photoID, err := s.svc.AddPhoto(ctx, id, userFromCtx(ctx).ID, ct, data)
	switch {
	case errors.Is(err, service.ErrNotFound):
		return oapi.UploadIncidentPhoto404JSONResponse(apiErr("not_found", "авария не найдена")), nil
	case errors.Is(err, service.ErrForbidden):
		return oapi.UploadIncidentPhoto403JSONResponse(apiErr("forbidden", "фото могут добавлять только автор и подписчики аварии")), nil
	case errors.Is(err, service.ErrInvalidStatus):
		return oapi.UploadIncidentPhoto422JSONResponse(apiErr("business_rule_failed", "к аварии можно прикрепить не больше 5 фото")), nil
	case errors.Is(err, service.ErrRateLimited):
		return oapi.UploadIncidentPhoto429JSONResponse(apiErr("rate_limited", "не больше 20 фото в час")), nil
	case err != nil:
		return nil, err
	}
	return oapi.UploadIncidentPhoto201JSONResponse(toPhoto(photoID)), nil
}

func (s *server) GetPhoto(ctx context.Context, req oapi.GetPhotoRequestObject) (oapi.GetPhotoResponseObject, error) {
	id, ok := parseID("ph_", req.Id)
	if !ok {
		return oapi.GetPhoto404JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("not_found", "фото не найдено"))}, nil
	}
	p, err := s.svc.GetPhoto(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.GetPhoto404JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("not_found", "фото не найдено"))}, nil
	}
	if err != nil {
		return nil, err
	}
	if p.ContentType == "image/png" {
		return oapi.GetPhoto200ImagepngResponse{Body: bytes.NewReader(p.Data), ContentLength: int64(len(p.Data))}, nil
	}
	return oapi.GetPhoto200ImagejpegResponse{Body: bytes.NewReader(p.Data), ContentLength: int64(len(p.Data))}, nil
}
