package service

import (
	"context"
	"errors"
	"time"

	"ukapp/internal/domain"
)

// UkProvider — порт к системе управляющей компании (contracts/uk.yaml). Единственная
// реализация в MVP — ukclient (HTTP к mock/uk); в production сюда встаёт адаптер
// настоящей УК-системы, service об этом не узнаёт.
type UkProvider interface {
	FindHouse(ctx context.Context, fiasID string) (UkHouse, error)
	RegisterIncident(ctx context.Context, in UkIncident) (id string, status domain.IncidentStatus, err error)
	IncidentUpdates(ctx context.Context, since time.Time) ([]UkIncidentUpdate, error)
	SetStatus(ctx context.Context, id string, status domain.IncidentStatus) error
}

var ErrUkHouseNotFound = errors.New("house not served by uk")

type UkHouse struct {
	ID, Address, OrgID, OrgName string
}

type UkIncident struct {
	ExternalRef, HouseID, Title, Description, Severity string
	Entrance, Riser                                    *string
}

type UkIncidentUpdate struct {
	ID, ExternalRef string
	Status          domain.IncidentStatus
	UpdatedAt       time.Time
}
