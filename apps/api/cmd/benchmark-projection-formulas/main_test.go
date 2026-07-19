package main

import (
	"bytes"
	"testing"
)

func TestParseOptionsRequiresInput(t *testing.T) {
	_, err := parseOptions(nil, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected missing input error")
	}
}

func TestParseOptionsAcceptsOutput(t *testing.T) {
	options, err := parseOptions(
		[]string{
			"--input",
			"request.json",
			"--output",
			"report.json",
		},
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if options.InputPath != "request.json" ||
		options.OutputPath != "report.json" {
		t.Fatalf("options = %+v", options)
	}
}

func TestRunRejectsMissingInputFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(
		[]string{
			"--input",
			"does-not-exist.json",
		},
		&stdout,
		&stderr,
	)
	if exitCode != exitInvalid {
		t.Fatalf(
			"exit code = %d, want %d",
			exitCode,
			exitInvalid,
		)
	}
}
