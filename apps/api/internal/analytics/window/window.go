package window

import "time"

type Window struct {
	start time.Time
	end   time.Time
}

func New(
	start time.Time,
	end time.Time,
) Window {
	if end.Before(start) {
		start, end = end, start
	}

	return Window{
		start: start,
		end:   end,
	}
}

func (w Window) Start() time.Time {
	return w.start
}

func (w Window) End() time.Time {
	return w.end
}

func (w Window) Duration() time.Duration {
	return w.end.Sub(w.start)
}

func (w Window) Contains(
	t time.Time,
) bool {
	return !t.Before(w.start) && !t.After(w.end)
}
