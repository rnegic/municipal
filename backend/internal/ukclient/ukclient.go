package ukclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"ukapp/internal/domain"
	"ukapp/internal/service"

	ukapi "ukapp/gen/uk"
)

type Client struct {
	api *ukapi.ClientWithResponses
}

func New(baseURL, token string) *Client {
	c, err := ukapi.NewClientWithResponses(baseURL,
		ukapi.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
		ukapi.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		}))
	if err != nil {
		panic(err)
	}
	return &Client{api: c}
}

func toDomainStatus(s ukapi.IncidentStatus) domain.IncidentStatus { return domain.IncidentStatus(s) }

func (c *Client) FindHouse(ctx context.Context, fiasID string) (service.UkHouse, error) {
	resp, err := c.api.FindHouseWithResponse(ctx, fiasID)
	if err != nil {
		return service.UkHouse{}, err
	}
	switch resp.StatusCode() {
	case http.StatusOK:
		h := resp.JSON200
		if h == nil {
			return service.UkHouse{}, fmt.Errorf("uk: find house: status %d without JSON body", resp.StatusCode())
		}
		return service.UkHouse{ID: h.Id, Address: h.Address, OrgID: h.Organization.Id, OrgName: h.Organization.Name}, nil
	case http.StatusNotFound:
		return service.UkHouse{}, service.ErrUkHouseNotFound
	default:
		return service.UkHouse{}, fmt.Errorf("uk: find house: status %d", resp.StatusCode())
	}
}

func (c *Client) RegisterIncident(ctx context.Context, in service.UkIncident) (string, domain.IncidentStatus, error) {
	resp, err := c.api.RegisterIncidentWithResponse(ctx, ukapi.RegisterIncidentJSONRequestBody{
		ExternalRef: in.ExternalRef, HouseId: in.HouseID, Title: in.Title, Description: in.Description,
		Severity: ukapi.Severity(in.Severity), Entrance: in.Entrance, Riser: in.Riser,
	})
	if err != nil {
		return "", "", err
	}
	var inc *ukapi.Incident
	switch resp.StatusCode() {
	case http.StatusCreated:
		inc = resp.JSON201
	case http.StatusOK:
		inc = resp.JSON200
	default:
		return "", "", fmt.Errorf("uk: register incident: status %d", resp.StatusCode())
	}
	if inc == nil {
		return "", "", fmt.Errorf("uk: register incident: status %d without JSON body", resp.StatusCode())
	}
	return inc.Id, toDomainStatus(inc.Status), nil
}

func (c *Client) IncidentUpdates(ctx context.Context, since time.Time) ([]service.UkIncidentUpdate, error) {
	resp, err := c.api.ListIncidentUpdatesWithResponse(ctx, &ukapi.ListIncidentUpdatesParams{UpdatedSince: since})
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("uk: list updates: status %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		return nil, fmt.Errorf("uk: list updates: status %d without JSON body", resp.StatusCode())
	}
	out := make([]service.UkIncidentUpdate, len(resp.JSON200.Items))
	for i, it := range resp.JSON200.Items {
		out[i] = service.UkIncidentUpdate{ID: it.Id, ExternalRef: it.ExternalRef, Status: toDomainStatus(it.Status), UpdatedAt: it.UpdatedAt}
	}
	return out, nil
}

func (c *Client) SetStatus(ctx context.Context, id string, status domain.IncidentStatus) error {
	resp, err := c.api.SetIncidentStatusWithResponse(ctx, id, ukapi.SetIncidentStatusJSONRequestBody{Status: ukapi.IncidentStatus(status)})
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("uk: set status: status %d", resp.StatusCode())
	}
	return nil
}
