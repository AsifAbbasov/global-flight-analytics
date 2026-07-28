package projectioncontract

import (
	"errors"
	"fmt"
)

var ErrResultInvalid = errors.New(
	"projection contract result is invalid",
)

type ResultValidationError struct {
	Issues []ValidationIssue
}

func (validationError *ResultValidationError) Error() string {
	if validationError == nil {
		return ErrResultInvalid.Error()
	}
	return fmt.Sprintf(
		"%s: issue_count=%d",
		ErrResultInvalid,
		len(validationError.Issues),
	)
}

func (validationError *ResultValidationError) Unwrap() error {
	return ErrResultInvalid
}

func (validationError *ResultValidationError) Clone() *ResultValidationError {
	if validationError == nil {
		return nil
	}
	return &ResultValidationError{
		Issues: append(
			[]ValidationIssue(nil),
			validationError.Issues...,
		),
	}
}

func (result Result) Validate() error {
	report := Validate(result)
	if report.Status == ValidationStatusValid {
		return nil
	}
	return &ResultValidationError{
		Issues: append(
			[]ValidationIssue(nil),
			report.Issues...,
		),
	}
}
