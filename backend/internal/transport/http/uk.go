package http

import (
	"context"
	"strings"

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

func (s *server) UkQueue(ctx context.Context, req oapi.UkQueueRequestObject) (oapi.UkQueueResponseObject, error) {
	offset, limit := int64(0), int64(20)
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
			Title: r.Title, Description: r.Description, Severity: oapi.Severity(r.Severity), Status: oapi.IncidentStatus(r.Status),
			CreatedAt: r.CreatedAt, DueAt: r.DueAt, AffectedCount: r.Subscribers, ConfirmedCount: r.Confirmations,
			ReporterName: shortName(r.ReporterName),
		}
	}
	return oapi.UkQueue200JSONResponse{Items: items, Total: int(total), Offset: int(offset), Limit: int(limit)}, nil
}

// shortName: "Иван Иванов" → "Иван И." (surname reduced to an initial for the dispatcher list).
func shortName(full string) string {
	first, last, ok := strings.Cut(strings.TrimSpace(full), " ")
	if !ok || last == "" {
		return first
	}
	return first + " " + string([]rune(last)[:1]) + "."
}
