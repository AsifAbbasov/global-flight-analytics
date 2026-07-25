package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFeaturePipelineConfigIgnoresFormatting(
	t *testing.T,
) {
	path := filepath.Join(t.TempDir(), "contracts.go")
	content := `package featurepipeline

import "example.test/flightfeatures"

type FeatureWriter interface {
	Put()
}

type Config struct {
	Extractor int
	Writer FeatureWriter
	ProcessingVersion flightfeatures.ProcessingVersion
}
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if failures := validateFeaturePipelineConfig(path); len(failures) != 0 {
		t.Fatalf(
			"validateFeaturePipelineConfig() failures = %#v",
			failures,
		)
	}
}

func TestValidateFeaturePipelineConfigRejectsWrongWriter(
	t *testing.T,
) {
	path := filepath.Join(t.TempDir(), "contracts.go")
	content := `package featurepipeline

import "example.test/flightfeatures"

type Config struct {
	Writer any
	ProcessingVersion flightfeatures.ProcessingVersion
}
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	failures := validateFeaturePipelineConfig(path)
	if len(failures) != 1 ||
		failures[0] !=
			"feature pipeline Config.Writer must use FeatureWriter" {
		t.Fatalf(
			"validateFeaturePipelineConfig() failures = %#v",
			failures,
		)
	}
}

func TestResolveRootAcceptsExplicitRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "apps", "api")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(path, "go.mod"),
		[]byte("module example.test/project\n"),
		0o644,
	); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := resolveRoot(root)
	if err != nil {
		t.Fatalf("resolveRoot() error = %v", err)
	}
	if got != root {
		t.Fatalf("resolveRoot() = %q, want %q", got, root)
	}
}
