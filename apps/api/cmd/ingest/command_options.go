package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

type ingestCommandOptions struct {
	Once bool
}

func parseIngestCommandOptions(
	args []string,
) (ingestCommandOptions, error) {
	flagSet := flag.NewFlagSet(
		"ingest",
		flag.ContinueOnError,
	)
	flagSet.SetOutput(io.Discard)

	once := flagSet.Bool(
		"once",
		false,
		"run exactly one ingestion cycle and exit",
	)

	if err := flagSet.Parse(args); err != nil {
		return ingestCommandOptions{}, fmt.Errorf(
			"parse ingest command options: %w",
			err,
		)
	}

	if flagSet.NArg() != 0 {
		return ingestCommandOptions{}, fmt.Errorf(
			"parse ingest command options: unexpected positional argument %q",
			strings.TrimSpace(flagSet.Arg(0)),
		)
	}

	return ingestCommandOptions{
		Once: *once,
	}, nil
}
