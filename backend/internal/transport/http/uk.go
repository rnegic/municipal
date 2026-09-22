package http

import (
	"context"

	oapi "ukapp/gen/api"
)

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
