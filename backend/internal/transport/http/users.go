package http

import (
	"context"
	"errors"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"

	"ukapp/gen/db/ukapp/public/model"
)

func toUser(u model.AppUser) oapi.User {
	return oapi.User{Id: formatUserID(u.ID), FullName: u.FullName, Role: oapi.Role(u.Role)}
}

func (s *server) GetMe(ctx context.Context, _ oapi.GetMeRequestObject) (oapi.GetMeResponseObject, error) {
	u := userFromCtx(ctx)
	resp := oapi.GetMe200JSONResponse{User: toUser(u)}
	if u.HouseID != nil {
		addr, err := s.svc.HouseAddress(ctx, *u.HouseID)
		if err != nil {
			return nil, err
		}
		resp.House = &oapi.House{Id: formatHouseID(*u.HouseID), Address: addr}
	}
	return resp, nil
}

func (s *server) BindHouse(ctx context.Context, req oapi.BindHouseRequestObject) (oapi.BindHouseResponseObject, error) {
	if req.Body.Address == "" {
		return oapi.BindHouse400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "address required"))}, nil
	}
	houseID, address, err := s.svc.BindHouse(ctx, userFromCtx(ctx).ID, req.Body.Address)
	if errors.Is(err, service.ErrAddressNotResolved) {
		return oapi.BindHouse422JSONResponse(apiErr("business_rule_failed", "не удалось распознать адрес")), nil
	}
	if errors.Is(err, service.ErrHouseNotServed) {
		return oapi.BindHouse422JSONResponse(apiErr("business_rule_failed", "дом не обслуживается подключённой УК")), nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.BindHouse200JSONResponse{Id: formatHouseID(houseID), Address: address}, nil
}

func (s *server) SuggestAddresses(ctx context.Context, req oapi.SuggestAddressesRequestObject) (oapi.SuggestAddressesResponseObject, error) {
	count := service.DefaultSuggestCount
	if req.Params.Count != nil {
		count = *req.Params.Count
	}
	sugs, err := s.svc.SuggestAddresses(ctx, req.Params.Query, count)
	if errors.Is(err, service.ErrInvalidInput) {
		return oapi.SuggestAddresses400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "query required"))}, nil
	}
	if err != nil {
		return oapi.SuggestAddresses502JSONResponse(apiErr("upstream_unavailable", "сервис подсказок адреса недоступен")), nil
	}
	items := make([]oapi.AddressSuggestion, 0, len(sugs))
	for _, sug := range sugs {
		items = append(items, oapi.AddressSuggestion{Value: sug.Value, HouseFiasId: sug.HouseFiasID})
	}
	return oapi.SuggestAddresses200JSONResponse{Suggestions: items}, nil
}
