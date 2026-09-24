package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"ukapp/internal/domain"
)

type blockingClassifier struct{}

func (blockingClassifier) Classify(ctx context.Context, _ string) (domain.Prediction, error) {
	<-ctx.Done()
	return domain.Prediction{}, ctx.Err()
}

type stubClassifier domain.Prediction

func (s stubClassifier) Classify(context.Context, string) (domain.Prediction, error) {
	return domain.Prediction(s), nil
}

func TestAnalyzeTimesOut(t *testing.T) {
	old := classifyTimeout
	classifyTimeout = 50 * time.Millisecond
	t.Cleanup(func() { classifyTimeout = old })
	s := New(nil, nil, nil, nil, "").WithClassifier(blockingClassifier{}, 0.7)
	start := time.Now()
	_, err := s.AnalyzeIncident(context.Background(), "лифт не работает третий день")
	if !errors.Is(err, ErrModelUnavailable) || time.Since(start) > time.Second {
		t.Fatalf("err=%v after %s", err, time.Since(start))
	}
}

func TestAnalyzeWithoutClassifier(t *testing.T) {
	_, err := New(nil, nil, nil, nil, "").AnalyzeIncident(context.Background(), "лифт не работает третий день")
	if !errors.Is(err, ErrModelUnavailable) {
		t.Fatal(err)
	}
}

func TestAnalyzeThresholdAndAuthorityTable(t *testing.T) {
	s := New(nil, nil, nil, nil, "").WithClassifier(stubClassifier{Category: domain.CategoryCityTerritory, P: 0.69}, 0.7)
	if _, err := s.AnalyzeIncident(context.Background(), "яма на дороге за двором"); !errors.Is(err, ErrModelUnavailable) {
		t.Fatalf("below threshold: %v", err)
	}
	s.WithClassifier(stubClassifier{Category: domain.CategoryCityTerritory, P: 0.7}, 0.7)
	a, err := s.AnalyzeIncident(context.Background(), "яма на дороге за двором")
	if err != nil || a.Authority != domain.AuthorityMunicipality || a.Reasoning != domain.AuthorityMunicipality.Reasoning() {
		t.Fatalf("%+v %v", a, err)
	}
	if _, err := s.AnalyzeIncident(context.Background(), "   коротко  "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("short: %v", err)
	}
}

func TestCreateIncidentFallsBackToManualWhenModelHangs(t *testing.T) {
	old := classifyTimeout
	classifyTimeout = 50 * time.Millisecond
	t.Cleanup(func() { classifyTimeout = old })
	s := testStore(t)
	svc := New(s, nil, nil, &fakeUk{}, "testbot").WithClassifier(blockingClassifier{}, 0.7)
	houseID, userID := seedHouse(t, s)
	start := time.Now()
	row, created, err := svc.CreateIncident(context.Background(), houseID, userID, NewIncident{Description: "лифт стоит третий день", Category: domain.CategoryElevator})
	if err != nil || !created || time.Since(start) > 2*time.Second {
		t.Fatalf("created=%v err=%v after %s", created, err, time.Since(start))
	}
	var src string
	if err := s.DB().QueryRowContext(context.Background(), `SELECT routing_source FROM incident WHERE id=$1`, row.ID).Scan(&src); err != nil || src != "manual" {
		t.Fatalf("routing_source=%q err=%v", src, err)
	}
}
