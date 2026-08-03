package main

import "testing"

func TestParseIngestCommandOptionsDefaultsToDaemonMode(
	t *testing.T,
) {
	t.Parallel()

	options, err := parseIngestCommandOptions(nil)
	if err != nil {
		t.Fatalf("parse default command options: %v", err)
	}
	if options.Once {
		t.Fatal("expected daemon mode by default")
	}
}

func TestParseIngestCommandOptionsAcceptsOnceMode(
	t *testing.T,
) {
	t.Parallel()

	options, err := parseIngestCommandOptions(
		[]string{"--once"},
	)
	if err != nil {
		t.Fatalf("parse one-shot command options: %v", err)
	}
	if !options.Once {
		t.Fatal("expected one-shot mode")
	}
}

func TestParseIngestCommandOptionsRejectsUnknownFlag(
	t *testing.T,
) {
	t.Parallel()

	if _, err := parseIngestCommandOptions(
		[]string{"--unknown"},
	); err == nil {
		t.Fatal("expected an unknown-flag error")
	}
}

func TestParseIngestCommandOptionsRejectsPositionalArgument(
	t *testing.T,
) {
	t.Parallel()

	if _, err := parseIngestCommandOptions(
		[]string{"unexpected"},
	); err == nil {
		t.Fatal("expected a positional-argument error")
	}
}
