package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/analytics/formulabenchmark"
)

const maximumInputBytes = 16 * 1024 * 1024

const (
	exitPassed               = 0
	exitInvalid              = 1
	exitInsufficientEvidence = 2
	exitBenchmarkFailed      = 3
)

type commandOptions struct {
	InputPath  string
	OutputPath string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	options, err := parseOptions(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitPassed
	}
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: parse options: %v\n",
			err,
		)
		return exitInvalid
	}

	request, err := readRequest(options.InputPath)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: read benchmark request: %v\n",
			err,
		)
		return exitInvalid
	}

	report, err :=
		formulabenchmark.Evaluate(request)
	if err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: evaluate benchmark request: %v\n",
			err,
		)
		return exitInvalid
	}

	if err := writeReport(
		options.OutputPath,
		stdout,
		report,
	); err != nil {
		fmt.Fprintf(
			stderr,
			"ERROR: write benchmark report: %v\n",
			err,
		)
		return exitInvalid
	}

	switch report.Status {
	case formulabenchmark.StatusBenchmarkPassed:
		return exitPassed
	case formulabenchmark.StatusInsufficientEvidence:
		return exitInsufficientEvidence
	case formulabenchmark.StatusBenchmarkFailed:
		return exitBenchmarkFailed
	default:
		return exitInvalid
	}
}

func parseOptions(
	args []string,
	stderr io.Writer,
) (commandOptions, error) {
	flags := flag.NewFlagSet(
		"benchmark-projection-formulas",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	inputPath := flags.String(
		"input",
		"",
		"path to a bounded offline benchmark request JSON file",
	)
	outputPath := flags.String(
		"output",
		"-",
		"report path or - for standard output",
	)

	if err := flags.Parse(args); err != nil {
		return commandOptions{}, err
	}
	if flags.NArg() != 0 {
		return commandOptions{}, fmt.Errorf(
			"unexpected positional arguments: %v",
			flags.Args(),
		)
	}

	normalizedInput := strings.TrimSpace(*inputPath)
	if normalizedInput == "" {
		return commandOptions{}, fmt.Errorf(
			"--input is required",
		)
	}

	normalizedOutput := strings.TrimSpace(*outputPath)
	if normalizedOutput == "" {
		return commandOptions{}, fmt.Errorf(
			"--output must not be empty",
		)
	}

	return commandOptions{
		InputPath:  normalizedInput,
		OutputPath: normalizedOutput,
	}, nil
}

func readRequest(
	path string,
) (
	formulabenchmark.Request,
	error,
) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return formulabenchmark.Request{}, err
	}
	defer file.Close()

	limited := io.LimitReader(
		file,
		maximumInputBytes+1,
	)
	content, err := io.ReadAll(limited)
	if err != nil {
		return formulabenchmark.Request{}, err
	}
	if len(content) > maximumInputBytes {
		return formulabenchmark.Request{}, fmt.Errorf(
			"input exceeds %d bytes",
			maximumInputBytes,
		)
	}

	var request formulabenchmark.Request
	decoder := json.NewDecoder(
		strings.NewReader(string(content)),
	)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return formulabenchmark.Request{}, err
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return formulabenchmark.Request{}, err
	}

	return request, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return fmt.Errorf(
			"input contains multiple JSON values",
		)
	}
	return err
}

func writeReport(
	path string,
	stdout io.Writer,
	report formulabenchmark.Report,
) error {
	if path == "-" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	}

	cleanPath := filepath.Clean(path)
	file, err := os.OpenFile(
		cleanPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o600,
	)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encodeErr := encoder.Encode(report)
	closeErr := file.Close()

	if encodeErr != nil {
		_ = os.Remove(cleanPath)
		return encodeErr
	}
	if closeErr != nil {
		_ = os.Remove(cleanPath)
		return closeErr
	}

	return nil
}
