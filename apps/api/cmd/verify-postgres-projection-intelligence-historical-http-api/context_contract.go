package main

import "errors"

var ErrVerificationContextRequired = errors.New(
	"Projection Intelligence verification context is required",
)
