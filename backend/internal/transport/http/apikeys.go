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
