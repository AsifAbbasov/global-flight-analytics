package handlers

import (
	"context"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routecontract"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routepipeline"
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/routeintelligence/routestore"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const handlerTrajectoryID = "8a3d6e20-2c68-4b35-a512-7d91e6a90c31"

type pipelineStub struct {
	result routepipeline.Result
	err    error
	id     string
}

func (s *pipelineStub) Process(_ context.Context, r routepipeline.Request) (routepipeline.Result, error) {
	s.id = r.TrajectoryID
	return s.result.Clone(), s.err
}

type storeStub struct {
	latest    routestore.Record
	latestErr error
	page      routestore.Page
	listErr   error
	query     routestore.ListQuery
}

func (s *storeStub) GetLatest(context.Context, string, routecontract.SchemaVersion) (routestore.Record, error) {
	return s.latest.Clone(), s.latestErr
}
func (s *storeStub) List(_ context.Context, q routestore.ListQuery) (routestore.Page, error) {
	s.query = q
	return s.page.Clone(), s.listErr
}

func TestRouteIntelligenceProcessEndpoint(t *testing.T) {
	record := handlerRecord()
	p := &pipelineStub{result: routepipeline.Result{Record: record}}
	h := NewRouteIntelligenceHandler(p, &storeStub{})
	app := fiber.New()
	app.Post("/api/v1/trajectories/:id/route-intelligence", h.ProcessByTrajectoryID)
	res, err := app.Test(httptest.NewRequest(http.MethodPost, "/api/v1/trajectories/"+handlerTrajectoryID+"/route-intelligence", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatalf("status %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), `"trajectory_id":"`+handlerTrajectoryID+`"`) {
		t.Fatalf("body %s", body)
	}
	if p.id != handlerTrajectoryID {
		t.Fatalf("id %q", p.id)
	}
}
func TestRouteIntelligenceHistoryParsing(t *testing.T) {
	s := &storeStub{page: routestore.Page{}}
	h := NewRouteIntelligenceHandler(&pipelineStub{}, s)
	app := fiber.New()
	app.Get("/api/v1/trajectories/:id/route-intelligence/history", h.ListHistoryByTrajectoryID)
	before := time.Date(2026, time.July, 14, 18, 0, 0, 0, time.UTC)
	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/trajectories/"+handlerTrajectoryID+"/route-intelligence/history?limit=7&before_as_of_time="+before.Format(time.RFC3339Nano), nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 || s.query.Limit != 7 || !s.query.BeforeAsOfTime.Equal(before) {
		t.Fatalf("status=%d query=%#v", res.StatusCode, s.query)
	}
}
func TestRouteIntelligenceInvalidIDAndNotFound(t *testing.T) {
	h := NewRouteIntelligenceHandler(&pipelineStub{}, &storeStub{latestErr: routestore.ErrResultNotFound})
	app := fiber.New()
	app.Post("/api/v1/trajectories/:id/route-intelligence", h.ProcessByTrajectoryID)
	app.Get("/api/v1/trajectories/:id/route-intelligence/latest", h.GetLatestByTrajectoryID)
	bad, _ := app.Test(httptest.NewRequest(http.MethodPost, "/api/v1/trajectories/bad/route-intelligence", nil))
	defer bad.Body.Close()
	if bad.StatusCode != 400 {
		t.Fatalf("bad status %d", bad.StatusCode)
	}
	missing, _ := app.Test(httptest.NewRequest(http.MethodGet, "/api/v1/trajectories/"+handlerTrajectoryID+"/route-intelligence/latest", nil))
	defer missing.Body.Close()
	if missing.StatusCode != 404 {
		t.Fatalf("missing status %d", missing.StatusCode)
	}
}
func TestRouteIntelligenceUnavailable(t *testing.T) {
	h := NewRouteIntelligenceHandler(nil, nil)
	app := fiber.New()
	app.Post("/:id", h.ProcessByTrajectoryID)
	res, _ := app.Test(httptest.NewRequest(http.MethodPost, "/"+handlerTrajectoryID, nil))
	defer res.Body.Close()
	if res.StatusCode != 503 {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func handlerRecord() routestore.Record {
	now := time.Date(2026, time.July, 14, 18, 0, 0, 0, time.UTC)
	return routestore.Record{ID: "route-record-test", InputFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", StoredAt: now, Key: routestore.ResultKey{TrajectoryID: handlerTrajectoryID, SchemaVersion: routecontract.SchemaVersionV1, AsOfTime: now}, Result: routecontract.Result{SchemaVersion: routecontract.SchemaVersionV1, Status: routecontract.RouteStatusUnavailable, TrajectoryID: handlerTrajectoryID, ICAO24: "ABC123", Window: routecontract.RouteWindow{StartTime: now.Add(-time.Hour), EndTime: now, AsOfTime: now}, Confidence: routecontract.Confidence{Level: routecontract.ConfidenceLevelNone}, Provenance: routecontract.Provenance{ResolverVersion: "route-resolver-v1", InputFingerprint: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", TrajectoryUpdatedAt: now, SourceNames: []string{"trajectory"}}, GeneratedAt: now}}
}
