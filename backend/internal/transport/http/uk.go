package http

import (
	"bytes"
	"context"
	"errors"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"
)

func (s *server) HouseSticker(ctx context.Context, req oapi.HouseStickerRequestObject) (oapi.HouseStickerResponseObject, error) {
	houseID, ok := parseID("h_", req.HouseId)
	if !ok {
		return oapi.HouseSticker404JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("not_found", "дом не найден"))}, nil
	}
	svg, err := s.svc.HouseSticker(ctx, houseID)
	if errors.Is(err, service.ErrNotFound) {
		return oapi.HouseSticker404JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("not_found", "дом не найден"))}, nil
	}
	if err != nil {
		return nil, err
	}
	return oapi.HouseSticker200ImagesvgXmlResponse{Body: bytes.NewReader(svg), ContentLength: int64(len(svg))}, nil
}

func (s *server) HouseStats(ctx context.Context, req oapi.HouseStatsRequestObject) (oapi.HouseStatsResponseObject, error) {
	houseID, ok := parseID("h_", req.HouseId)
	if ok {
		var err error
		if ok, err = s.svc.CanAccessHouse(ctx, userFromCtx(ctx), houseID); err != nil {
			return nil, err
		}
	}
	if !ok {
		return oapi.HouseStats404JSONResponse(apiErr("not_found", "дом не найден")), nil
	}
	st, err := s.svc.HouseStats(ctx, houseID)
	if err != nil {
		return nil, err
	}
	return oapi.HouseStats200JSONResponse{
		ActiveIncidents: st.ActiveIncidents, InProgress: st.InProgress, ResolvedLast30Days: st.ResolvedLast30Days,
		AvgResolutionHours: st.AvgResolutionHours, LastIncidentAt: st.LastIncidentAt,
	}, nil
}
