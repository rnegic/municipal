package gisgkh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"ukapp/internal/service"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func New() *Client {
	return &Client{
		BaseURL: "https://dom.gosuslugi.ru/homemanagement/api/rest/services/houses/public/searchByAddress",
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

type org struct {
	FullName   string  `json:"fullName"`
	ShortName  string  `json:"shortName"`
	OrgAddress *string `json:"orgAddress"`
	Phone      *string `json:"phone"`
	URL        *string `json:"url"`
	INN        string  `json:"inn"`
	OGRN       string  `json:"ogrn"`
}

func (c *Client) FindOrg(ctx context.Context, a service.HouseAddress) (*service.UkOrg, error) {
	if a.RegionFiasID == "" || a.StreetFiasID == "" {
		return nil, nil
	}
	body, _ := json.Marshal(map[string]any{
		"regionCode": a.RegionFiasID, "streetCode": a.StreetFiasID, "houseNumber": a.House, "calcCount": true,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"?pageIndex=1&elementsPerPage=20", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Session-GUID", uuid.NewString())
	req.Header.Set("Request-GUID", uuid.NewString())
	req.Header.Set("State-GUID", "/houses")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gisgkh: status %d", resp.StatusCode)
	}
	var out struct {
		Items []struct {
			Status  string `json:"status"`
			Address struct {
				House struct {
					HouseGUID string `json:"houseGuid"`
				} `json:"house"`
			} `json:"address"`
			ManagementOrganization *org `json:"managementOrganization"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	for _, it := range out.Items {
		o := it.ManagementOrganization
		if it.Status != "APPROVED" || o == nil || o.INN == "" || it.Address.House.HouseGUID != a.HouseFiasID {
			continue
		}
		name := o.ShortName
		if name == "" {
			name = o.FullName
		}
		return &service.UkOrg{
			ExternalID: "inn-" + o.INN, Name: name,
			Phone: o.Phone, Website: o.URL, OfficeAddress: o.OrgAddress,
		}, nil
	}
	return nil, nil
}
