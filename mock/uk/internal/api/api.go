// Package api — реализация контракта contracts/uk.yaml (gin strict-server из gen/api).
package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	ukapi "mockuk/gen/api"
	"mockuk/internal/lk"
	"mockuk/internal/store"
)

type server struct{ st *store.Store }

func validStatus(s ukapi.IncidentStatus) bool {
	switch s {
	case ukapi.Pending, ukapi.Accepted, ukapi.InProgress, ukapi.Verifying, ukapi.Done, ukapi.FalseAlarm:
		return true
	default:
		return false
	}
}

func NewRouter(st *store.Store, token string, failureRate float64) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	h := ukapi.NewStrictHandlerWithOptions(&server{st: st}, nil, ukapi.StrictGinServerOptions{
		RequestErrorHandlerFunc: func(c *gin.Context, err error) {
			writeErr(c, http.StatusBadRequest, "validation_failed", err.Error())
		},
		ResponseErrorHandlerFunc: func(c *gin.Context, err error) {
			slog.Error("response handler error", "err", err)
			writeErr(c, http.StatusInternalServerError, "internal", "internal error")
		},
	})
	ukapi.RegisterHandlersWithOptions(r, h, ukapi.GinServerOptions{
		Middlewares: []ukapi.MiddlewareFunc{bearer(token), faults(failureRate)},
	})
	lk.Mount(r, st)
	return r
}

func toWire(in store.Incident) ukapi.Incident {
	return ukapi.Incident{Id: in.ID, ExternalRef: in.ExternalRef, Status: ukapi.IncidentStatus(in.Status), UpdatedAt: in.UpdatedAt}
}

func (s *server) FindHouse(ctx context.Context, req ukapi.FindHouseRequestObject) (ukapi.FindHouseResponseObject, error) {
	h, err := s.st.FindHouseByFias(ctx, req.FiasId)
	if errors.Is(err, store.ErrNotFound) {
		return ukapi.FindHouse404JSONResponse{Code: "not_found", Message: "дом не обслуживается"}, nil
	}
	if err != nil {
		return nil, err
	}
	return ukapi.FindHouse200JSONResponse{
		Id: h.ID, Address: h.Address,
		Organization: ukapi.Organization{
			Id: h.OrgID, Name: h.OrgName,
			Phone: h.OrgPhone, EmergencyPhone: h.OrgEmergencyPhone, Email: h.OrgEmail,
			Website: h.OrgWebsite, OfficeAddress: h.OrgOfficeAddress, WorkingHours: h.OrgWorkingHours,
		},
	}, nil
}

func (s *server) RegisterIncident(ctx context.Context, req ukapi.RegisterIncidentRequestObject) (ukapi.RegisterIncidentResponseObject, error) {
	b := req.Body
	if b.Title == "" {
		return ukapi.RegisterIncident400JSONResponse{Code: "validation_failed", Message: "title required"}, nil
	}
	inc, created, err := s.st.CreateIncident(ctx, store.Incident{
		ExternalRef: b.ExternalRef, HouseID: b.HouseId, Title: b.Title, Description: b.Description,
		Severity: string(b.Severity), Entrance: b.Entrance, Riser: b.Riser, Suspicious: b.Suspicious != nil && *b.Suspicious,
	})
	if errors.Is(err, store.ErrNotFound) {
		return ukapi.RegisterIncident404JSONResponse{Code: "not_found", Message: "unknown houseId"}, nil
	}
	if err != nil {
		return nil, err
	}
	if created {
		return ukapi.RegisterIncident201JSONResponse(toWire(inc)), nil
	}
	return ukapi.RegisterIncident200JSONResponse(toWire(inc)), nil
}

func (s *server) ListIncidentUpdates(ctx context.Context, req ukapi.ListIncidentUpdatesRequestObject) (ukapi.ListIncidentUpdatesResponseObject, error) {
	items, err := s.st.UpdatedSince(ctx, req.Params.UpdatedSince)
	if err != nil {
		return nil, err
	}
	out := make([]ukapi.Incident, len(items))
	for i, in := range items {
		out[i] = toWire(in)
	}
	return ukapi.ListIncidentUpdates200JSONResponse{Items: out}, nil
}

func (s *server) SetIncidentStatus(ctx context.Context, req ukapi.SetIncidentStatusRequestObject) (ukapi.SetIncidentStatusResponseObject, error) {
	if !validStatus(req.Body.Status) {
		return ukapi.SetIncidentStatus400JSONResponse{Code: "validation_failed", Message: "unknown status"}, nil
	}
	inc, err := s.st.SetStatus(ctx, req.Id, string(req.Body.Status))
	switch {
	case errors.Is(err, store.ErrNotFound):
		return ukapi.SetIncidentStatus404JSONResponse{Code: "not_found", Message: "no such incident"}, nil
	case errors.Is(err, store.ErrConflict):
		return ukapi.SetIncidentStatus409JSONResponse{Code: "conflict", Message: "incident is closed"}, nil
	case err != nil:
		return nil, err
	}
	return ukapi.SetIncidentStatus200JSONResponse(toWire(inc)), nil
}
