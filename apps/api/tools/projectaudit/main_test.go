package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTypeScriptInterfaces(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"contract.ts",
	)
	content := `export interface Example {
  id: string
  count: number
}
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	interfaces, err := parseTypeScriptInterfaces(
		path,
	)
	if err != nil {
		t.Fatalf(
			"parse interfaces: %v",
			err,
		)
	}

	example := interfaces["Example"]
	if example.Fields["id"] != "string" ||
		example.Fields["count"] != "number" {
		t.Fatalf(
			"unexpected fields: %#v",
			example.Fields,
		)
	}
}

func TestParseTypeScriptStringUnions(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(
		directory,
		"contract.ts",
	)
	content := `export type Example =
  | 'alpha'
  | 'beta'

export interface Other {
  id: string
}
`
	if err := os.WriteFile(
		path,
		[]byte(content),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	unions, err := parseTypeScriptStringUnions(
		path,
	)
	if err != nil {
		t.Fatalf(
			"parse unions: %v",
			err,
		)
	}

	values := unions["Example"]
	if len(values) != 2 ||
		values[0] != "alpha" ||
		values[1] != "beta" {
		t.Fatalf(
			"values = %#v",
			values,
		)
	}
}
