package http

import (
	"context"
	"errors"
	"strconv"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"

	"ukapp/gen/db/ukapp/public/model"
)

func formatApiKeyID(id int64) string { return "key_" + strconv.FormatInt(id, 10) }

func toApiKey(k model.UkAPIKey) oapi.ApiKey {
	return oapi.ApiKey{
		Id: formatApiKeyID(k.ID), Name: k.Name, Prefix: k.Prefix,
		CreatedAt: k.CreatedAt, LastUsedAt: k.LastUsedAt, RevokedAt: k.RevokedAt,
	}
}

func (s *server) CreateUkApiKey(ctx context.Context, req oapi.CreateUkApiKeyRequestObject) (oapi.CreateUkApiKeyResponseObject, error) {
	created, err := s.svc.CreateApiKey(ctx, userFromCtx(ctx), req.Body.Name)
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return oapi.CreateUkApiKey400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "нужно название ключа до 100 символов"))}, nil
	case errors.Is(err, service.ErrKeyLimit):
		return oapi.CreateUkApiKey422JSONResponse(apiErr("business_rule_failed", "у организации уже 10 активных ключей, отзовите лишние")), nil
	case err != nil:
		return nil, err
	}
	k := toApiKey(created.Key)
	return oapi.CreateUkApiKey201JSONResponse{
		Id: k.Id, Name: k.Name, Prefix: k.Prefix, CreatedAt: k.CreatedAt,
		LastUsedAt: k.LastUsedAt, RevokedAt: k.RevokedAt, Key: created.Plain,
	}, nil
}

func (s *server) ListUkApiKeys(ctx context.Context, _ oapi.ListUkApiKeysRequestObject) (oapi.ListUkApiKeysResponseObject, error) {
	keys, err := s.svc.ListApiKeys(ctx, userFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	items := make([]oapi.ApiKey, len(keys))
	for i, k := range keys {
		items[i] = toApiKey(k)
	}
	return oapi.ListUkApiKeys200JSONResponse{Items: items}, nil
}

func (s *server) RevokeUkApiKey(ctx context.Context, req oapi.RevokeUkApiKeyRequestObject) (oapi.RevokeUkApiKeyResponseObject, error) {
	notFound := oapi.RevokeUkApiKey404JSONResponse(apiErr("not_found", "ключ не найден"))
	id, ok := parseID("key_", req.Id)
	if !ok {
		return notFound, nil
	}
	err := s.svc.RevokeApiKey(ctx, userFromCtx(ctx), id)
	if errors.Is(err, service.ErrNotFound) {
		return notFound, nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.RevokeUkApiKey204Response{}, nil
}

func (s *server) ListUkIncidentChanges(ctx context.Context, req oapi.ListUkIncidentChangesRequestObject) (oapi.ListUkIncidentChangesResponseObject, error) {
	cursor, limit := int64(0), int64(100)
	if req.Params.Cursor != nil {
		c, err := strconv.ParseInt(*req.Params.Cursor, 10, 64)
		if err != nil || c < 0 {
			return oapi.ListUkIncidentChanges400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "cursor — неотрицательное целое из nextCursor"))}, nil
		}
		cursor = c
	}
	if req.Params.Limit != nil {
		if *req.Params.Limit < 1 || *req.Params.Limit > 500 {
			return oapi.ListUkIncidentChanges400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "limit — от 1 до 500"))}, nil
		}
		limit = int64(*req.Params.Limit)
	}
	rows, err := s.svc.UkChanges(ctx, userFromCtx(ctx), cursor, limit)
	if err != nil {
		return nil, err
	}
	items := make([]oapi.UkChangeItem, len(rows))
	for i, r := range rows {
		var mergedInto *string
		if r.MergedIntoID != nil {
			m := formatIncidentID(*r.MergedIntoID)
			mergedInto = &m
		}
		items[i] = oapi.UkChangeItem{
			Id: formatIncidentID(r.ID), HouseId: formatHouseID(r.HouseID), HouseAddress: r.HouseAddress,
			Title: r.Title, Description: r.Description, Severity: oapi.Severity(r.Severity), Status: oapi.IncidentStatus(r.Status),
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, AffectedCount: r.AffectedCount, ConfirmedCount: r.ConfirmedCount,
			ReporterName: r.ReporterName, Photos: toPhotos(r.PhotoIDs),
			Category: (*oapi.IncidentCategory)(r.Category), MergedCount: r.MergedCount,
			MergedIntoId: mergedInto, UpdatedAt: r.UpdatedAt,
		}
		cursor = r.ChangeSeq
	}
	return oapi.ListUkIncidentChanges200JSONResponse{Items: items, NextCursor: strconv.FormatInt(cursor, 10)}, nil
}
