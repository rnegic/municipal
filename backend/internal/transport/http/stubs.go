package http

// Code stubs for operations not implemented yet (P1 per docs/frontend-api-contract);
// delete a method here when implementing it.

import (
	"context"

	oapi "ukapp/gen/api"
)

func (s *server) HouseStats(context.Context, oapi.HouseStatsRequestObject) (oapi.HouseStatsResponseObject, error) {
	return nil, errNotImplemented
}

func (s *server) UploadIncidentPhoto(context.Context, oapi.UploadIncidentPhotoRequestObject) (oapi.UploadIncidentPhotoResponseObject, error) {
	return nil, errNotImplemented
}
