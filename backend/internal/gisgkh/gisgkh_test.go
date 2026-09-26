package gisgkh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ukapp/internal/service"
)

const searchResp = `{"total":3,"items":[
{"status":"CANCELLED","address":{"house":{"houseGuid":"f-1"}},"managementOrganization":{"shortName":"Старая","inn":"1"}},
{"status":"APPROVED","address":{"house":{"houseGuid":"f-other"}},"managementOrganization":{"shortName":"Соседняя","inn":"2"}},
{"status":"APPROVED","address":{"house":{"houseGuid":"f-1"}},"managementOrganization":{"fullName":"ООО УК ВАХИТОВСКОГО РАЙОНА","shortName":"ООО \"УК Вахитовского района\"","phone":"78432000000","url":"https://uk.ru","orgAddress":"Казань","inn":"1655000001","ogrn":"1"},"municipalityOrganization":{"shortName":"МКУ","inn":"9"}}
]}`

func fake(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b map[string]any
		_ = json.NewDecoder(r.Body).Decode(&b)
		if b["regionCode"] != "reg" || b["streetCode"] != "str" || b["houseNumber"] != "7/10" || r.Header.Get("Session-GUID") == "" {
			t.Errorf("bad request %v %v", b, r.Header)
		}
		_, _ = w.Write([]byte(searchResp))
	}))
	t.Cleanup(srv.Close)
	c := New()
	c.BaseURL = srv.URL
	return c
}

func TestFindOrg(t *testing.T) {
	c := fake(t)
	org, err := c.FindOrg(context.Background(), service.HouseAddress{HouseFiasID: "f-1", RegionFiasID: "reg", StreetFiasID: "str", House: "7/10"})
	if err != nil || org == nil {
		t.Fatalf("org=%v err=%v", org, err)
	}
	if org.ExternalID != "inn-1655000001" || org.Name != `ООО "УК Вахитовского района"` || *org.Phone != "78432000000" || *org.Website != "https://uk.ru" {
		t.Fatalf("got %+v", org)
	}
}

func TestFindOrg_NoApproved(t *testing.T) {
	c := fake(t)
	org, err := c.FindOrg(context.Background(), service.HouseAddress{HouseFiasID: "f-none", RegionFiasID: "reg", StreetFiasID: "str", House: "7/10"})
	if err != nil || org != nil {
		t.Fatalf("want nil, got %v %v", org, err)
	}
}
