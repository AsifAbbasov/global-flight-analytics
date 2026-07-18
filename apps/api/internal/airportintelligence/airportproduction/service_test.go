package airportproduction

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/airport"
)

func TestServiceExercisesCompleteAirportIntelligenceFlow(t *testing.T) {
	asOfTime := time.Date(2026, time.July, 18, 12, 0, 0, 0, time.UTC)
	windowStart := time.Date(2026, time.July, 16, 0, 0, 0, 0, time.UTC)
	repository := fakeAirportRepository{items: []airport.Airport{
		{ICAOCode: "UBBB", IATACode: "GYD", Name: "Baku", City: "Baku", Country: "Azerbaijan", Latitude: 40.4675, Longitude: 50.0467},
		{ICAOCode: "UGTB", IATACode: "TBS", Name: "Tbilisi", City: "Tbilisi", Country: "Georgia", Latitude: 41.6692, Longitude: 44.9547},
	}}
	reader := fakeObservationReader{items: []DailyObservation{
		dailyObservation("UBBB", windowStart, 10, 8, 7, 3),
		dailyObservation("UBBB", windowStart.Add(24*time.Hour), 20, 16, 9, 4),
		dailyObservation("UGTB", windowStart, 4, 3, 3, 2),
		dailyObservation("UGTB", windowStart.Add(24*time.Hour), 5, 4, 4, 2),
	}}
	service, err := New(Config{AirportRepository: repository, ObservationReader: reader, Now: func() time.Time { return asOfTime }})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}
	request := WindowRequest{AsOfTime: asOfTime, Days: 2}

	overviewResult, err := service.GetOverview(context.Background(), "ubbb", request)
	if err != nil {
		t.Fatalf("get overview: %v", err)
	}
	if overviewResult.Overview.Passport.Identity.ICAOCode != "UBBB" {
		t.Fatalf("overview ICAO = %q", overviewResult.Overview.Passport.Identity.ICAOCode)
	}
	if overviewResult.Overview.Statistics.TotalMovements != 54 {
		t.Fatalf("total movements = %d, want 54", overviewResult.Overview.Statistics.TotalMovements)
	}
	if overviewResult.Overview.Ranking.Position != 1 {
		t.Fatalf("ranking position = %d, want 1", overviewResult.Overview.Ranking.Position)
	}

	historyResult, err := service.GetHistory(context.Background(), "UBBB", request)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(historyResult.History.Entries) != 2 {
		t.Fatalf("history entries = %d, want 2", len(historyResult.History.Entries))
	}

	trendsResult, err := service.GetTrends(context.Background(), "UBBB", request)
	if err != nil {
		t.Fatalf("get trends: %v", err)
	}
	if string(trendsResult.Trends.Direction) != "increasing" {
		t.Fatalf("trend direction = %q", trendsResult.Trends.Direction)
	}

	rankingResult, err := service.GetRanking(context.Background(), request)
	if err != nil {
		t.Fatalf("get ranking: %v", err)
	}
	if len(rankingResult.Ranking.Airports) != 2 {
		t.Fatalf("ranking airports = %d, want 2", len(rankingResult.Ranking.Airports))
	}
	if rankingResult.Ranking.Airports[0].ICAOCode != "UBBB" {
		t.Fatalf("first ranked ICAO = %q", rankingResult.Ranking.Airports[0].ICAOCode)
	}
}

func TestServiceRejectsWindowOutsidePolicy(t *testing.T) {
	service, err := New(Config{AirportRepository: fakeAirportRepository{}, ObservationReader: fakeObservationReader{}, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.GetRanking(context.Background(), WindowRequest{Days: MaximumWindowDays + 1})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidRequest)
	}
}

func TestNewPostgresObservationReaderRejectsNilPool(t *testing.T) {
	_, err := NewPostgresObservationReader(nil)
	if !errors.Is(err, ErrPostgresPoolRequired) {
		t.Fatalf("error = %v, want %v", err, ErrPostgresPoolRequired)
	}
}

type fakeAirportRepository struct{ items []airport.Airport }

func (repository fakeAirportRepository) List(context.Context) ([]airport.Airport, error) {
	return append([]airport.Airport(nil), repository.items...), nil
}
func (repository fakeAirportRepository) GetByICAO(_ context.Context, icaoCode string) (airport.Airport, error) {
	normalized := strings.ToUpper(strings.TrimSpace(icaoCode))
	for _, item := range repository.items {
		if strings.ToUpper(strings.TrimSpace(item.ICAOCode)) == normalized {
			return item, nil
		}
	}
	return airport.Airport{}, airport.ErrNotFound
}

type fakeObservationReader struct{ items []DailyObservation }

func (reader fakeObservationReader) ListDaily(_ context.Context, query DailyQuery) ([]DailyObservation, error) {
	result := make([]DailyObservation, 0)
	for _, item := range reader.items {
		if query.ICAOCode != "" && item.ICAOCode != query.ICAOCode {
			continue
		}
		if item.WindowStart.Before(query.WindowStart) || item.WindowEnd.After(query.WindowEnd) {
			continue
		}
		result = append(result, item)
	}
	return result, nil
}

func dailyObservation(icaoCode string, windowStart time.Time, arrivals, departures, activeAircraft, activeRoutes int) DailyObservation {
	windowEnd := windowStart.Add(24 * time.Hour)
	return DailyObservation{ICAOCode: icaoCode, WindowStart: windowStart, WindowEnd: windowEnd, Arrivals: arrivals, Departures: departures, ActiveAircraft: activeAircraft, ActiveRoutes: activeRoutes, ObservedAt: windowEnd.Add(-time.Microsecond)}
}
