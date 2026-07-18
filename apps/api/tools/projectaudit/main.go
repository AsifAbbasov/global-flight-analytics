package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const modulePath = "github.com/AsifAbbasov/global-flight-analytics/apps/api"

type auditMode string

const (
	modeAll          auditMode = "all"
	modeReachability auditMode = "reachability"
	modeContracts    auditMode = "contracts"
	modeDuplicates   auditMode = "duplicates"
)

type packageInfo struct {
	ImportPath   string
	Name         string
	Dir          string
	Imports      []string
	TestImports  []string
	XTestImports []string
}

type rootSpec struct {
	Name    string
	Pattern string
}

type contextSpec struct {
	Name                 string
	RequiredRuntimeReach bool
	AcceptedRuntimeRoots map[string]struct{}
}

type contextResult struct {
	Name                string
	Total               int
	RuntimeReachable    int
	VerificationOnly    int
	NotRuntimeReachable []string
	RootCounts          map[string]int
}

type goField struct {
	Name string
	Type string
}

type tsInterface struct {
	Name   string
	Fields map[string]string
}

func main() {
	os.Exit(run(
		os.Args[1:],
		os.Stdout,
		os.Stderr,
	))
}

func run(
	args []string,
	stdout io.Writer,
	stderr io.Writer,
) int {
	flags := flag.NewFlagSet(
		"projectaudit",
		flag.ContinueOnError,
	)
	flags.SetOutput(stderr)

	modeValue := flags.String(
		"mode",
		string(modeAll),
		"audit mode: all, reachability, contracts, or duplicates",
	)
	strict := flags.Bool(
		"strict",
		true,
		"fail when a required runtime analytical context is unreachable",
	)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	mode := auditMode(
		strings.TrimSpace(
			*modeValue,
		),
	)
	if !knownMode(mode) {
		fmt.Fprintf(
			stderr,
			"unknown audit mode %q\n",
			mode,
		)
		return 1
	}

	repositoryRoot, err := findRepositoryRoot()
	if err != nil {
		fmt.Fprintf(
			stderr,
			"locate repository root: %v\n",
			err,
		)
		return 1
	}

	failures := make([]string, 0)

	if mode == modeAll ||
		mode == modeDuplicates {
		if err := auditConfidenceDuplication(
			repositoryRoot,
			stdout,
		); err != nil {
			failures = append(
				failures,
				err.Error(),
			)
		}
	}

	if mode == modeAll ||
		mode == modeContracts {
		if err := auditTrajectoryContract(
			repositoryRoot,
			stdout,
		); err != nil {
			failures = append(
				failures,
				err.Error(),
			)
		}
	}

	if mode == modeAll ||
		mode == modeReachability {
		if err := auditReachability(
			repositoryRoot,
			*strict,
			stdout,
		); err != nil {
			failures = append(
				failures,
				err.Error(),
			)
		}
	}

	if len(failures) > 0 {
		fmt.Fprintln(
			stderr,
			"Project architecture audit: FAIL",
		)
		for _, failure := range failures {
			fmt.Fprintf(
				stderr,
				"- %s\n",
				failure,
			)
		}
		return 1
	}

	fmt.Fprintln(
		stdout,
		"Project architecture audit: PASS",
	)
	return 0
}

func knownMode(
	mode auditMode,
) bool {
	switch mode {
	case modeAll,
		modeReachability,
		modeContracts,
		modeDuplicates:
		return true
	default:
		return false
	}
}

func findRepositoryRoot() (
	string,
	error,
) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goMod := filepath.Join(
			current,
			"apps",
			"api",
			"go.mod",
		)
		trajectoryType := filepath.Join(
			current,
			"apps",
			"web",
			"types",
			"trajectory.ts",
		)
		if regularFile(goMod) &&
			regularFile(trajectoryType) {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New(
				"repository root containing apps/api/go.mod and apps/web/types/trajectory.ts was not found",
			)
		}
		current = parent
	}
}

func regularFile(
	path string,
) bool {
	info, err := os.Stat(path)
	return err == nil &&
		info.Mode().IsRegular()
}

func auditConfidenceDuplication(
	repositoryRoot string,
	output io.Writer,
) error {
	requiredFragments := map[string][]string{
		filepath.Join(
			repositoryRoot,
			"apps/api/internal/domain/dataquality/model.go",
		): {
			`type ConfidenceLevel = domainconfidence.Level`,
			`domainconfidence.LevelHigh`,
			`domainconfidence.LevelMedium`,
			`domainconfidence.LevelLow`,
			`domainconfidence.LevelNone`,
		},
		filepath.Join(
			repositoryRoot,
			"apps/api/internal/domain/metrics/model.go",
		): {
			`type ConfidenceLevel = domainconfidence.Level`,
			`domainconfidence.LevelHigh`,
			`domainconfidence.LevelMedium`,
			`domainconfidence.LevelLow`,
			`domainconfidence.LevelNone`,
		},
	}

	for path, fragments := range requiredFragments {
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf(
				"read confidence consumer %s: %w",
				path,
				err,
			)
		}
		content := string(contentBytes)

		if strings.Contains(
			content,
			"type ConfidenceLevel string",
		) {
			return fmt.Errorf(
				"duplicated string-backed ConfidenceLevel remains in %s",
				path,
			)
		}

		for _, fragment := range fragments {
			if !strings.Contains(
				content,
				fragment,
			) {
				return fmt.Errorf(
					"confidence consumer %s does not reference the shared value object: missing %q",
					path,
					fragment,
				)
			}
		}
	}

	fmt.Fprintln(
		output,
		"Confidence vocabulary audit: PASS",
	)
	return nil
}

func auditTrajectoryContract(
	repositoryRoot string,
	output io.Writer,
) error {
	goDTOPath := filepath.Join(
		repositoryRoot,
		"apps/api/internal/http/dto/trajectory.go",
	)
	goDomainPath := filepath.Join(
		repositoryRoot,
		"apps/api/internal/domain/trajectory/model.go",
	)
	tsPath := filepath.Join(
		repositoryRoot,
		"apps/web/types/trajectory.ts",
	)

	goStructs, err := parseGoStructs(
		goDTOPath,
	)
	if err != nil {
		return err
	}
	tsInterfaces, err := parseTypeScriptInterfaces(
		tsPath,
	)
	if err != nil {
		return err
	}

	structMappings := map[string]string{
		"Trajectory":        "AircraftTrajectory",
		"TrajectorySegment": "TrajectorySegment",
		"CoverageGap":       "CoverageGap",
	}
	typeOverrides := map[string]string{
		"AircraftTrajectory.identity_basis": "FlightIdentityBasis",
		"AircraftTrajectory.split_reason":   "FlightSplitReason",
		"TrajectorySegment.status":          "TrajectorySegmentStatus",
		"CoverageGap.reason":                "CoverageGapReason",
	}

	for goName, tsName := range structMappings {
		goFields, ok := goStructs[goName]
		if !ok {
			return fmt.Errorf(
				"Go DTO struct %s is missing",
				goName,
			)
		}
		tsValue, ok := tsInterfaces[tsName]
		if !ok {
			return fmt.Errorf(
				"TypeScript interface %s is missing",
				tsName,
			)
		}

		if err := compareContractFields(
			goName,
			goFields,
			tsValue,
			typeOverrides,
		); err != nil {
			return err
		}
	}

	goEnums, err := parseGoStringEnums(
		goDomainPath,
	)
	if err != nil {
		return err
	}
	tsEnums, err := parseTypeScriptStringUnions(
		tsPath,
	)
	if err != nil {
		return err
	}

	enumMappings := map[string]string{
		"SegmentStatus":       "TrajectorySegmentStatus",
		"CoverageGapReason":   "CoverageGapReason",
		"FlightIdentityBasis": "FlightIdentityBasis",
		"FlightSplitReason":   "FlightSplitReason",
	}
	for goName, tsName := range enumMappings {
		if err := compareStringSets(
			goName,
			goEnums[goName],
			tsName,
			tsEnums[tsName],
		); err != nil {
			return err
		}
	}

	if err := auditTrajectoryRuntimeParser(
		repositoryRoot,
		goEnums,
	); err != nil {
		return err
	}

	fmt.Fprintln(
		output,
		"Go and TypeScript trajectory contract audit: PASS",
	)
	return nil
}

func parseGoStructs(
	path string,
) (
	map[string]map[string]goField,
	error,
) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		parser.ParseComments,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse Go DTO file %s: %w",
			path,
			err,
		)
	}

	result := make(
		map[string]map[string]goField,
	)

	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok ||
			general.Tok != token.TYPE {
			continue
		}

		for _, specification := range general.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			fields := make(
				map[string]goField,
			)
			for _, field := range structType.Fields.List {
				if field.Tag == nil ||
					len(field.Names) == 0 {
					continue
				}

				rawTag, err := strconv.Unquote(
					field.Tag.Value,
				)
				if err != nil {
					return nil, fmt.Errorf(
						"unquote struct tag in %s.%s: %w",
						typeSpec.Name.Name,
						field.Names[0].Name,
						err,
					)
				}
				jsonName := jsonTagName(rawTag)
				if jsonName == "" ||
					jsonName == "-" {
					continue
				}

				fieldType, err := goTypeToTypeScript(
					field.Type,
				)
				if err != nil {
					return nil, fmt.Errorf(
						"map Go field type %s.%s: %w",
						typeSpec.Name.Name,
						field.Names[0].Name,
						err,
					)
				}

				fields[jsonName] = goField{
					Name: jsonName,
					Type: fieldType,
				}
			}

			result[typeSpec.Name.Name] = fields
		}
	}

	return result, nil
}

func jsonTagName(
	rawTag string,
) string {
	for _, part := range strings.Fields(
		rawTag,
	) {
		if !strings.HasPrefix(
			part,
			`json:"`,
		) {
			continue
		}
		value := strings.TrimPrefix(
			part,
			`json:"`,
		)
		value = strings.TrimSuffix(
			value,
			`"`,
		)
		return strings.SplitN(
			value,
			",",
			2,
		)[0]
	}
	return ""
}

func goTypeToTypeScript(
	expression ast.Expr,
) (
	string,
	error,
) {
	switch value := expression.(type) {
	case *ast.Ident:
		switch value.Name {
		case "string":
			return "string", nil
		case "int",
			"int8",
			"int16",
			"int32",
			"int64",
			"uint",
			"uint8",
			"uint16",
			"uint32",
			"uint64",
			"float32",
			"float64":
			return "number", nil
		case "bool":
			return "boolean", nil
		default:
			return value.Name, nil
		}
	case *ast.SelectorExpr:
		if identifier, ok := value.X.(*ast.Ident); ok &&
			identifier.Name == "time" &&
			value.Sel.Name == "Time" {
			return "string", nil
		}
		return value.Sel.Name, nil
	case *ast.ArrayType:
		element, err := goTypeToTypeScript(
			value.Elt,
		)
		if err != nil {
			return "", err
		}
		return element + "[]", nil
	case *ast.StarExpr:
		element, err := goTypeToTypeScript(
			value.X,
		)
		if err != nil {
			return "", err
		}
		return element + " | null", nil
	default:
		return "", fmt.Errorf(
			"unsupported Go expression %T",
			expression,
		)
	}
}

func parseTypeScriptInterfaces(
	path string,
) (
	map[string]tsInterface,
	error,
) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"read TypeScript contract %s: %w",
			path,
			err,
		)
	}

	startPattern := regexp.MustCompile(
		`^\s*export\s+interface\s+([A-Za-z0-9_]+)\s*\{\s*$`,
	)
	fieldPattern := regexp.MustCompile(
		`^\s*([A-Za-z0-9_]+)\??\s*:\s*(.+?)\s*$`,
	)

	result := make(
		map[string]tsInterface,
	)
	scanner := bufio.NewScanner(
		bytes.NewReader(contentBytes),
	)

	currentName := ""
	currentFields := make(
		map[string]string,
	)

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)
		if currentName == "" {
			match := startPattern.FindStringSubmatch(
				line,
			)
			if len(match) == 2 {
				currentName = match[1]
				currentFields = make(
					map[string]string,
				)
			}
			continue
		}

		if line == "}" {
			result[currentName] = tsInterface{
				Name:   currentName,
				Fields: currentFields,
			}
			currentName = ""
			currentFields = make(
				map[string]string,
			)
			continue
		}

		if line == "" ||
			strings.HasPrefix(line, "//") {
			continue
		}

		match := fieldPattern.FindStringSubmatch(
			line,
		)
		if len(match) != 3 {
			return nil, fmt.Errorf(
				"unsupported TypeScript interface line in %s: %q",
				currentName,
				line,
			)
		}

		currentFields[match[1]] = strings.TrimSpace(
			strings.TrimSuffix(
				match[2],
				";",
			),
		)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if currentName != "" {
		return nil, fmt.Errorf(
			"unterminated TypeScript interface %s",
			currentName,
		)
	}

	return result, nil
}

func compareContractFields(
	goName string,
	goFields map[string]goField,
	tsValue tsInterface,
	overrides map[string]string,
) error {
	missing := make(
		[]string,
		0,
	)
	extra := make(
		[]string,
		0,
	)
	typeMismatches := make(
		[]string,
		0,
	)

	for name, goFieldValue := range goFields {
		tsType, exists := tsValue.Fields[name]
		if !exists {
			missing = append(
				missing,
				name,
			)
			continue
		}

		expectedType := goFieldValue.Type
		if override, ok := overrides[tsValue.Name+"."+name]; ok {
			expectedType = override
		}

		if normalizeTypeScriptType(tsType) !=
			normalizeTypeScriptType(expectedType) {
			typeMismatches = append(
				typeMismatches,
				fmt.Sprintf(
					"%s: Go=%s TypeScript=%s",
					name,
					goFieldValue.Type,
					tsType,
				),
			)
		}
	}

	for name := range tsValue.Fields {
		if _, exists := goFields[name]; !exists {
			extra = append(
				extra,
				name,
			)
		}
	}

	sort.Strings(missing)
	sort.Strings(extra)
	sort.Strings(typeMismatches)

	if len(missing) > 0 ||
		len(extra) > 0 ||
		len(typeMismatches) > 0 {
		return fmt.Errorf(
			"Go DTO %s and TypeScript interface %s drifted: missing=%v extra=%v type_mismatches=%v",
			goName,
			tsValue.Name,
			missing,
			extra,
			typeMismatches,
		)
	}

	return nil
}

func normalizeTypeScriptType(
	value string,
) string {
	return strings.Join(
		strings.Fields(value),
		" ",
	)
}

func parseGoStringEnums(
	path string,
) (
	map[string][]string,
	error,
) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(
		fileSet,
		path,
		nil,
		0,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse Go domain file %s: %w",
			path,
			err,
		)
	}

	result := make(
		map[string][]string,
	)

	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok ||
			general.Tok != token.CONST {
			continue
		}

		for _, specification := range general.Specs {
			valueSpec, ok := specification.(*ast.ValueSpec)
			if !ok ||
				valueSpec.Type == nil ||
				len(valueSpec.Values) != 1 {
				continue
			}

			typeIdentifier, ok := valueSpec.Type.(*ast.Ident)
			if !ok {
				continue
			}
			literal, ok := valueSpec.Values[0].(*ast.BasicLit)
			if !ok ||
				literal.Kind != token.STRING {
				continue
			}
			value, err := strconv.Unquote(
				literal.Value,
			)
			if err != nil {
				return nil, err
			}
			result[typeIdentifier.Name] = append(
				result[typeIdentifier.Name],
				value,
			)
		}
	}

	for name := range result {
		sort.Strings(
			result[name],
		)
	}

	return result, nil
}

func parseTypeScriptStringUnions(
	path string,
) (
	map[string][]string,
	error,
) {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	startPattern := regexp.MustCompile(
		`^\s*export\s+type\s+([A-Za-z0-9_]+)\s*=\s*(.*)$`,
	)
	valuePattern := regexp.MustCompile(
		`'([^']+)'`,
	)

	result := make(
		map[string][]string,
	)
	scanner := bufio.NewScanner(
		bytes.NewReader(contentBytes),
	)

	currentName := ""

	for scanner.Scan() {
		line := strings.TrimSpace(
			scanner.Text(),
		)

		if match := startPattern.FindStringSubmatch(
			line,
		); len(match) == 3 {
			currentName = match[1]
			for _, valueMatch := range valuePattern.FindAllStringSubmatch(
				match[2],
				-1,
			) {
				result[currentName] = append(
					result[currentName],
					valueMatch[1],
				)
			}
			continue
		}

		if currentName == "" {
			continue
		}

		if strings.HasPrefix(
			line,
			"|",
		) {
			for _, valueMatch := range valuePattern.FindAllStringSubmatch(
				line,
				-1,
			) {
				result[currentName] = append(
					result[currentName],
					valueMatch[1],
				)
			}
			continue
		}

		currentName = ""
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for name := range result {
		sort.Strings(
			result[name],
		)
	}

	return result, nil
}

func compareStringSets(
	goName string,
	goValues []string,
	tsName string,
	tsValues []string,
) error {
	if len(goValues) == 0 {
		return fmt.Errorf(
			"Go enum %s is missing or empty",
			goName,
		)
	}
	if len(tsValues) == 0 {
		return fmt.Errorf(
			"TypeScript union %s is missing or empty",
			tsName,
		)
	}

	if strings.Join(goValues, "\x00") !=
		strings.Join(tsValues, "\x00") {
		return fmt.Errorf(
			"Go enum %s and TypeScript union %s drifted: Go=%v TypeScript=%v",
			goName,
			tsName,
			goValues,
			tsValues,
		)
	}

	return nil
}

func auditReachability(
	repositoryRoot string,
	strict bool,
	output io.Writer,
) error {
	apiRoot := filepath.Join(
		repositoryRoot,
		"apps",
		"api",
	)

	packages, err := loadPackages(
		apiRoot,
	)
	if err != nil {
		return err
	}

	runtimeRoots := []rootSpec{
		{
			Name:    "server",
			Pattern: "./cmd/server",
		},
		{
			Name:    "ingest",
			Pattern: "./cmd/ingest",
		},
		{
			Name:    "reconcile",
			Pattern: "./cmd/reconcile",
		},
		{
			Name:    "historical_materializer",
			Pattern: "./cmd/materialize-historical-intelligence",
		},
	}

	rootDependencies := make(
		map[string]map[string]struct{},
	)
	runtimeUnion := make(
		map[string]struct{},
	)

	for _, root := range runtimeRoots {
		dependencies, err := listDependencies(
			apiRoot,
			root.Pattern,
		)
		if err != nil {
			return fmt.Errorf(
				"load runtime root %s: %w",
				root.Name,
				err,
			)
		}
		rootDependencies[root.Name] = dependencies
		for dependency := range dependencies {
			runtimeUnion[dependency] = struct{}{}
		}
	}

	verificationUnion := make(
		map[string]struct{},
	)
	for _, packageValue := range packages {
		if packageValue.Name != "main" {
			continue
		}
		if !strings.Contains(
			packageValue.ImportPath,
			"/cmd/verify-",
		) &&
			!strings.Contains(
				packageValue.ImportPath,
				"/cmd/audit-",
			) &&
			!strings.Contains(
				packageValue.ImportPath,
				"/tools/projectaudit",
			) {
			continue
		}

		pattern := "./" + strings.TrimPrefix(
			packageValue.ImportPath,
			modulePath+"/",
		)
		dependencies, err := listDependencies(
			apiRoot,
			pattern,
		)
		if err != nil {
			return fmt.Errorf(
				"load verification root %s: %w",
				packageValue.ImportPath,
				err,
			)
		}
		for dependency := range dependencies {
			verificationUnion[dependency] = struct{}{}
		}
	}

	contexts := []contextSpec{
		{
			Name:                 "analytics",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "airportintelligence",
			RequiredRuntimeReach: false,
		},
		{
			Name:                 "airspaceintelligence",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "features",
			RequiredRuntimeReach: false,
		},
		{
			Name:                 "historicalintelligence",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "projectionintelligence",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "routeintelligence",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "stabilityintelligence",
			RequiredRuntimeReach: true,
		},
		{
			Name:                 "weatherintelligence",
			RequiredRuntimeReach: true,
		},
	}

	results := make(
		[]contextResult,
		0,
		len(contexts),
	)
	failures := make(
		[]string,
		0,
	)

	for _, context := range contexts {
		prefix := modulePath +
			"/internal/" +
			context.Name

		result := contextResult{
			Name:       context.Name,
			RootCounts: make(map[string]int),
		}

		for _, packageValue := range packages {
			if packageValue.ImportPath != prefix &&
				!strings.HasPrefix(
					packageValue.ImportPath,
					prefix+"/",
				) {
				continue
			}

			result.Total++

			if _, reachable := runtimeUnion[packageValue.ImportPath]; reachable {
				result.RuntimeReachable++
				for rootName, dependencies := range rootDependencies {
					if _, rootReachable := dependencies[packageValue.ImportPath]; rootReachable {
						result.RootCounts[rootName]++
					}
				}
				continue
			}

			verificationReachable := false
			if _, verificationOnly := verificationUnion[packageValue.ImportPath]; verificationOnly {
				result.VerificationOnly++
				verificationReachable = true
			}

			if _, classified := nonRuntimePackagePolicyFor(
				packageValue.ImportPath,
			); !classified {
				failures = append(
					failures,
					fmt.Sprintf(
						"non-runtime package %s has no explicit disposition policy; verification_reachable=%t",
						packageValue.ImportPath,
						verificationReachable,
					),
				)
			}

			result.NotRuntimeReachable = append(
				result.NotRuntimeReachable,
				packageValue.ImportPath,
			)
		}

		sort.Strings(
			result.NotRuntimeReachable,
		)

		if result.Total == 0 {
			failures = append(
				failures,
				fmt.Sprintf(
					"context %s has no Go packages",
					context.Name,
				),
			)
		}
		if strict &&
			context.RequiredRuntimeReach &&
			result.RuntimeReachable == 0 {
			failures = append(
				failures,
				fmt.Sprintf(
					"required analytical context %s is not reachable from any runtime root",
					context.Name,
				),
			)
		}

		results = append(
			results,
			result,
		)
	}

	fmt.Fprintln(
		output,
		"Analytical production reachability",
	)
	for _, result := range results {
		fmt.Fprintf(
			output,
			"- %s: total=%d runtime=%d verification_only=%d server=%d ingest=%d reconcile=%d historical_materializer=%d\n",
			result.Name,
			result.Total,
			result.RuntimeReachable,
			result.VerificationOnly,
			result.RootCounts["server"],
			result.RootCounts["ingest"],
			result.RootCounts["reconcile"],
			result.RootCounts["historical_materializer"],
		)

		for _, packagePath := range result.NotRuntimeReachable {
			policy, classified :=
				nonRuntimePackagePolicyFor(
					packagePath,
				)
			if !classified {
				continue
			}

			_, verificationReachable :=
				verificationUnion[packagePath]

			fmt.Fprintf(
				output,
				"  CLASSIFIED_NOT_RUNTIME_REACHABLE package=%s disposition=%s verification_reachable=%t rationale=%q next_action=%q\n",
				packagePath,
				policy.Disposition,
				verificationReachable,
				policy.Rationale,
				policy.NextAction,
			)
		}
	}

	if len(failures) > 0 {
		return errors.New(
			strings.Join(
				failures,
				"; ",
			),
		)
	}

	fmt.Fprintln(
		output,
		"Analytical production reachability audit: PASS",
	)
	return nil
}

func loadPackages(
	apiRoot string,
) (
	[]packageInfo,
	error,
) {
	command := exec.Command(
		"go",
		"list",
		"-json",
		"./...",
	)
	command.Dir = apiRoot
	output, err := command.Output()
	if err != nil {
		return nil, commandError(
			command,
			err,
		)
	}

	decoder := json.NewDecoder(
		bytes.NewReader(output),
	)
	packages := make(
		[]packageInfo,
		0,
	)

	for {
		var packageValue packageInfo
		err := decoder.Decode(
			&packageValue,
		)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(
				"decode go list package stream: %w",
				err,
			)
		}
		if strings.HasPrefix(
			packageValue.ImportPath,
			modulePath,
		) {
			packages = append(
				packages,
				packageValue,
			)
		}
	}

	return packages, nil
}

func listDependencies(
	apiRoot string,
	pattern string,
) (
	map[string]struct{},
	error,
) {
	command := exec.Command(
		"go",
		"list",
		"-deps",
		"-f",
		"{{.ImportPath}}",
		pattern,
	)
	command.Dir = apiRoot
	output, err := command.Output()
	if err != nil {
		return nil, commandError(
			command,
			err,
		)
	}

	result := make(
		map[string]struct{},
	)
	scanner := bufio.NewScanner(
		bytes.NewReader(output),
	)
	for scanner.Scan() {
		value := strings.TrimSpace(
			scanner.Text(),
		)
		if strings.HasPrefix(
			value,
			modulePath,
		) {
			result[value] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func commandError(
	command *exec.Cmd,
	commandErr error,
) error {
	exitError, ok := commandErr.(*exec.ExitError)
	if !ok {
		return commandErr
	}

	return fmt.Errorf(
		"%w: %s",
		commandErr,
		strings.TrimSpace(
			string(exitError.Stderr),
		),
	)
}

// STAGE-14-1-TRAJECTORY-RUNTIME-PARSER-FIX

// STAGE-14-2-DEAD-CODE-CLASSIFICATION
