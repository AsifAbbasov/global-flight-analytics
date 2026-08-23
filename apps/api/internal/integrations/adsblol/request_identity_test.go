package adsblol

import "testing"

func TestIdentifiedUserAgentAddsPublicContactURL(t *testing.T) {
	got := identifiedUserAgent("global-flight-analytics-ingest")
	want := "global-flight-analytics-ingest (+https://github.com/AsifAbbasov/global-flight-analytics)"

	if got != want {
		t.Fatalf("identified user agent = %q, want %q", got, want)
	}
}

func TestIdentifiedUserAgentDoesNotDuplicateContactURL(t *testing.T) {
	input := "global-flight-analytics-ingest (+https://github.com/AsifAbbasov/global-flight-analytics)"

	if got := identifiedUserAgent(input); got != input {
		t.Fatalf("identified user agent = %q, want %q", got, input)
	}
}
