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
		return oapi.AnalyzeIncident503JSONResponse(apiErr("model_unavailable", "не удалось определить категорию, выберите вручную")), nil
	case err != nil:
		return nil, err
	}
	return oapi.AnalyzeIncident200JSONResponse{
		Category: oapi.IncidentCategory(a.Category), Authority: oapi.IncidentAuthority(a.Authority),
		IsUkResponsibility: a.Authority == domain.AuthorityUK, PhotoRequired: a.Category.PhotoRequired(),
		ReasoningText: &a.Reasoning,
	}, nil
}
