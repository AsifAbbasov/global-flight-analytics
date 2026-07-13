package executor

import (
	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/calculator"
)

type Executor struct {
	calculator *calculator.Calculator
}

func New(
	calculator *calculator.Calculator,
) *Executor {
	return &Executor{
		calculator: calculator,
	}
}

func (e *Executor) Calculator() *calculator.Calculator {
	return e.calculator
}
