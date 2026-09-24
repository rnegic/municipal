package service

import (
	"context"
	"os"
	"testing"

	"ukapp/internal/domain"
	"ukapp/internal/repository"
)

func seedHouse(t *testing.T, s *repository.Store) (houseID, userID int64) {
	t.Helper()
	ctx := context.Background()
	u, err := s.UpsertUser(ctx, domain.InitUser{ID: 200, FirstName: "R"})
	if err != nil {
		t.Fatal(err)
	}
	ukID, err := s.UpsertUk(ctx, "uk-1", "Демо УК")
	if err != nil {
		t.Fatal(err)
	}
	houseID, err = s.BindHouse(ctx, u.ID, "Баумана 10", "f-10", ukID, "h-f-10")
	if err != nil {
		t.Fatal(err)
	}
	return houseID, u.ID
}

type svcReport struct {
	incidentID, reporterID int64
	outcome, dedupVersion  string
}

func serviceReports(t *testing.T, s *repository.Store) []svcReport {
	t.Helper()
	rows, err := s.DB().QueryContext(context.Background(),
		`SELECT incident_id, reporter_id, outcome, dedup_version FROM incident_report ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []svcReport
	for rows.Next() {
		var r svcReport
		if err := rows.Scan(&r.incidentID, &r.reporterID, &r.outcome, &r.dedupVersion); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestCreateIncident_ReportsCreated(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	houseID, userID := seedHouse(t, s)
	row, created, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	got := serviceReports(t, s)
	if len(got) != 1 || got[0].incidentID != row.ID || got[0].reporterID != userID ||
		got[0].outcome != "created" || got[0].dedupVersion != "exact-category-floorzone-v1" {
		t.Fatalf("unexpected reports: %+v", got)
	}
}

func TestCreateIncident_ReportsJoined(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	houseID, userID := seedHouse(t, s)
	first, _, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.UpsertUser(context.Background(), domain.InitUser{ID: 201, FirstName: "S"})
	if err != nil {
		t.Fatal(err)
	}
	row, created, err := svc.CreateIncident(context.Background(), houseID, other.ID, NewIncident{Description: "нет воды и у нас", Category: domain.CategoryWaterHeat})
	if err != nil || created || row.ID != first.ID {
		t.Fatalf("want join into %d, got id=%d created=%v err=%v", first.ID, row.ID, created, err)
	}
	got := serviceReports(t, s)
	if len(got) != 2 || got[1].incidentID != first.ID || got[1].reporterID != other.ID || got[1].outcome != "joined" {
		t.Fatalf("unexpected reports: %+v", got)
	}
}

func TestCreateIncident_RetryDoesNotDuplicateReport(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	houseID, userID := seedHouse(t, s)
	for range 2 {
		if _, _, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat}); err != nil {
			t.Fatal(err)
		}
	}
	got := serviceReports(t, s)
	if len(got) != 1 || got[0].outcome != "created" {
		t.Fatalf("retry must not add a report, got %+v", got)
	}
}

func TestCreateIncident_ReportFailureDoesNotBreakRequest(t *testing.T) {
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot")
	houseID, userID := seedHouse(t, s)
	if _, err := s.DB().ExecContext(context.Background(), `DROP TABLE incident_report`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		st, err := repository.Open(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			t.Error(err)
			return
		}
		st.Close()
	})
	row, created, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "нет воды с утра", Category: domain.CategoryWaterHeat})
	if err != nil || !created || row.ID == 0 {
		t.Fatalf("request must succeed without report table: id=%d created=%v err=%v", row.ID, created, err)
	}
}
