package window

import (
	"testing"
	"time"
)

func TestNewSwapsDates(
	t *testing.T,
) {
	end := time.Now()

	start := end.Add(-time.Hour)

	window := New(end, start)

	if window.Start() != start {
		t.Fatal("unexpected start")
	}

	if window.End() != end {
		t.Fatal("unexpected end")
	}
}

func TestDuration(
	t *testing.T,
) {
	start := time.Now()

	end := start.Add(2 * time.Hour)

	window := New(start, end)

	if window.Duration() != 2*time.Hour {
		t.Fatal("unexpected duration")
	}
}

func TestContains(
	t *testing.T,
) {
	start := time.Now()

	end := start.Add(time.Hour)

	window := New(start, end)

	if !window.Contains(start.Add(30 * time.Minute)) {
		t.Fatal("time must be inside window")
	}

	if window.Contains(end.Add(time.Second)) {
		t.Fatal("time must be outside window")
	}
}
