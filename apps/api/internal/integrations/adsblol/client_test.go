package adsblol

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	integrationcommon "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/common"
)

func TestClientGetByPoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		request *http.Request,
	) {
		if request.URL.Path != "/v2/point/40.409300/49.867100/250" {
			t.Fatalf("path = %s", request.URL.Path)
		}
		if request.Header.Get("User-Agent") != "gfa-test" {
			t.Fatalf("user-agent = %q", request.Header.Get("User-Agent"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(
			`{"ac":[],"now":1675633671226,"total":0,"ctime":1675633671226,"ptime":0,"msg":"No error"}`,
		))
	}))
	defer server.Close()

	client, err := NewClient(integrationcommon.HTTPClientConfig{
		BaseURL:   server.URL,
		Timeout:   time.Second,
		UserAgent: "gfa-test",
	})
	if err != nil {
		t.Fatal(err)
	}

	response, err := client.GetByPoint(
		context.Background(),
		40.4093,
		49.8671,
		250,
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Now != 1675633671226 {
		t.Fatalf("now = %d", response.Now)
	}
}

func TestClientRejectsRadiusOver250(t *testing.T) {
	client, err := NewClient(integrationcommon.HTTPClientConfig{
		BaseURL:   "https://example.test",
		Timeout:   time.Second,
		UserAgent: "gfa-test",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetByPoint(
		context.Background(),
		40,
		49,
		251,
	); err == nil {
		t.Fatal("expected radius rejection")
	}
}
