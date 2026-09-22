package repository

import (
	"context"
	"testing"

	"ukapp/internal/domain"
)

func seedIncident(t *testing.T, s *Store) (houseID, userID, incidentID int64) {
	t.Helper()
	ctx := context.Background()
	u, err := s.UpsertUser(ctx, domain.InitUser{ID: 100, FirstName: "R"})
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
	incidentID, err = s.CreateIncident(ctx, houseID, u.ID, "Нет воды", "с утра", domain.SeverityCritical, nil, nil, domain.SLA(domain.SeverityCritical))
	if err != nil {
		t.Fatal(err)
	}
	return houseID, u.ID, incidentID
}

type reportRow struct {
	incidentID, reporterID, houseID                     int64
	title, description, severity, outcome, dedupVersion string
	entrance, riser                                     *string
}

func reports(t *testing.T, s *Store) []reportRow {
	t.Helper()
	rows, err := s.db.QueryContext(context.Background(),
		`SELECT incident_id, reporter_id, house_id, title, description, severity, entrance, riser, outcome, dedup_version FROM incident_report ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []reportRow
	for rows.Next() {
		var r reportRow
		if err := rows.Scan(&r.incidentID, &r.reporterID, &r.houseID, &r.title, &r.description, &r.severity, &r.entrance, &r.riser, &r.outcome, &r.dedupVersion); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestAddReport_SavesAllFields(t *testing.T) {
	s := testStore(t)
	houseID, userID, incID := seedIncident(t, s)
	entrance, riser := "2", "Б"
	err := s.AddReport(context.Background(), ReportInput{
		IncidentID: incID, ReporterID: userID, HouseID: houseID,
		Title: "Нет воды", Description: "с утра", Severity: domain.SeverityWarning,
		Entrance: &entrance, Riser: &riser,
		Outcome: domain.ReportCreated, DedupVersion: domain.DedupVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := reports(t, s)
	if len(got) != 1 {
		t.Fatalf("want 1 report, got %d", len(got))
	}
	r := got[0]
	if r.incidentID != incID || r.reporterID != userID || r.houseID != houseID ||
		r.title != "Нет воды" || r.description != "с утра" || r.severity != "warning" ||
		r.entrance == nil || *r.entrance != "2" || r.riser == nil || *r.riser != "Б" ||
		r.outcome != "created" || r.dedupVersion != "exact-title-riser-v1" {
		t.Fatalf("unexpected row: %+v", r)
	}
}

func TestAddReport_Joined(t *testing.T) {
	s := testStore(t)
	houseID, _, incID := seedIncident(t, s)
	other, err := s.UpsertUser(context.Background(), domain.InitUser{ID: 101, FirstName: "S"})
	if err != nil {
		t.Fatal(err)
	}
	err = s.AddReport(context.Background(), ReportInput{
		IncidentID: incID, ReporterID: other.ID, HouseID: houseID,
		Title: "Нет воды", Description: "тоже", Severity: domain.SeverityCritical,
		Outcome: domain.ReportJoined, DedupVersion: domain.DedupVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := reports(t, s)
	if len(got) != 1 || got[0].outcome != "joined" || got[0].reporterID != other.ID {
		t.Fatalf("unexpected rows: %+v", got)
	}
}

func TestAddReport_NullableFields(t *testing.T) {
	s := testStore(t)
	houseID, userID, incID := seedIncident(t, s)
	err := s.AddReport(context.Background(), ReportInput{
		IncidentID: incID, ReporterID: userID, HouseID: houseID,
		Title: "Нет воды", Description: "с утра", Severity: domain.SeverityCritical,
		Outcome: domain.ReportCreated, DedupVersion: domain.DedupVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	got := reports(t, s)
	if len(got) != 1 || got[0].entrance != nil || got[0].riser != nil {
		t.Fatalf("want NULL entrance/riser, got %+v", got)
	}
}

func TestAddReport_DuplicateIsIgnored(t *testing.T) {
	s := testStore(t)
	houseID, userID, incID := seedIncident(t, s)
	in := ReportInput{
		IncidentID: incID, ReporterID: userID, HouseID: houseID,
		Title: "Нет воды", Description: "с утра", Severity: domain.SeverityCritical,
		Outcome: domain.ReportCreated, DedupVersion: domain.DedupVersion,
	}
	for range 2 {
		if err := s.AddReport(context.Background(), in); err != nil {
			t.Fatal(err)
		}
	}
	if got := reports(t, s); len(got) != 1 {
		t.Fatalf("want 1 report after repeat, got %d", len(got))
	}
}
