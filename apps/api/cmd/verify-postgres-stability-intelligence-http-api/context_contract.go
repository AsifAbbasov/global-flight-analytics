package main

import "errors"

var ErrRuntimeStabilityContextRequired = errors.New(
	"Stability Intelligence verification context is required",
)
