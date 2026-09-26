package http

import (
	"context"
	"errors"

	oapi "ukapp/gen/api"
	"ukapp/internal/domain"
	"ukapp/internal/service"
)

func (s *server) AnalyzeIncident(ctx context.Context, req oapi.AnalyzeIncidentRequestObject) (oapi.AnalyzeIncidentResponseObject, error) {
	a, err := s.svc.AnalyzeIncident(ctx, req.Body.Description)
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return oapi.AnalyzeIncident400JSONResponse{ErrorJSONResponse: oapi.ErrorJSONResponse(apiErr("validation_failed", "описание: от 10 до 2000 символов"))}, nil
	case errors.Is(err, service.ErrModelUnavailable):
		return oapi.AnalyzeIncident200JSONResponse{}, nil
	case err != nil:
		return nil, err
	}
	cat, auth := oapi.IncidentCategory(a.Category), oapi.IncidentAuthority(a.Authority)
	return oapi.AnalyzeIncident200JSONResponse{
		Category: &cat, Authority: &auth,
		IsUkResponsibility: a.Authority == domain.AuthorityUK, PhotoRequired: a.Category.PhotoRequired(),
		ReasoningText: &a.Reasoning,
	}, nil
}
