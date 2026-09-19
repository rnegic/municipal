package http

// Code stubs for operations not implemented yet (P1/P2 per docs/frontend-api-contract);
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

func (s *server) SetIncidentStatus(context.Context, oapi.SetIncidentStatusRequestObject) (oapi.SetIncidentStatusResponseObject, error) {
	return nil, errNotImplemented
}

func (s *server) ListEvents(context.Context, oapi.ListEventsRequestObject) (oapi.ListEventsResponseObject, error) {
	return nil, errNotImplemented
}

func (s *server) GetEvent(context.Context, oapi.GetEventRequestObject) (oapi.GetEventResponseObject, error) {
	return nil, errNotImplemented
}

func (s *server) UkQueue(context.Context, oapi.UkQueueRequestObject) (oapi.UkQueueResponseObject, error) {
	return nil, errNotImplemented
}

func (s *server) UkCreateEvent(context.Context, oapi.UkCreateEventRequestObject) (oapi.UkCreateEventResponseObject, error) {
	return nil, errNotImplemented
}
