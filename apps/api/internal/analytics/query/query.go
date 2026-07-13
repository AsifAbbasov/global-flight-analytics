package query

import "time"

type Request struct {
	Metric string
	From   time.Time
	To     time.Time
}

func New(metric string, from, to time.Time) Request {
	return Request{
		Metric: metric,
		From:   from.UTC(),
		To:     to.UTC(),
	}
}

func (r Request) Duration() time.Duration {
	return r.To.Sub(r.From)
}

func (r Request) IsValid() bool {
	if r.Metric == "" {
		return false
	}

	return !r.To.Before(r.From)
}
