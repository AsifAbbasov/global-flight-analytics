package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const auditVersion = "backend-context-ownership-audit-v1"

type finding struct {
	Path      string
	Line      int
	Parameter string
	Root      string
}

type auditReport struct {
	FilesScanned      int
	ContextParameters int
	Findings          []finding
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args[1:]))
}

func run(stdout io.Writer, stderr io.Writer, args []string) int {
	flags := flag.NewFlagSet("backendcontextownershipaudit", flag.ContinueOnError)
	flags.SetOutput(stderr)

	root := flags.String("root", ".", "Go module root to audit")
	strict := flags.Bool("strict", false, "fail when findings exist")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	report, err := scanRepository(*root)
	if err != nil {
		fmt.Fprintf(stderr, "backend context ownership audit: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "Audit version: %s\n", auditVersion)
	fmt.Fprintf(stdout, "Production Go files scanned: %d\n", report.FilesScanned)
	fmt.Fprintf(stdout, "Context parameters inspected: %d\n", report.ContextParameters)

	if len(report.Findings) == 0 {
		fmt.Fprintln(stdout, "Caller context replacements: 0")
		fmt.Fprintln(stdout, "Backend caller context ownership audit: PASS")
		return 0
	}

	fmt.Fprintf(stdout, "Caller context replacements: %d\n", len(report.Findings))
	fmt.Fprintln(stdout, "Backend caller context ownership audit: FAIL")
	for _, item := range report.Findings {
		fmt.Fprintf(
			stdout,
			"- %s:%d context parameter %q is replaced through context.%s\n",
			item.Path,
			item.Line,
			item.Parameter,
			item.Root,
		)
	}

	if *strict {
		return 1
	}
	return 0
}

func scanRepository(root string) (auditReport, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return auditReport{}, fmt.Errorf("resolve root %q: %w", root, err)
	}
	if _, err := os.Stat(filepath.Join(absoluteRoot, "go.mod")); err != nil {
		return auditReport{}, fmt.Errorf("Go module root %q is invalid: %w", absoluteRoot, err)
	}

	report := auditReport{}
	fileSet := token.NewFileSet()
	for _, topLevel := range []string{"cmd", "internal", "tools"} {
		start := filepath.Join(absoluteRoot, topLevel)
		info, statErr := os.Stat(start)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return auditReport{}, fmt.Errorf("inspect audit root %q: %w", start, statErr)
		}
		if !info.IsDir() {
			continue
		}

		walkErr := filepath.WalkDir(
			start,
			func(path string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() {
					switch entry.Name() {
					case "testdata", "vendor":
						return filepath.SkipDir
					default:
						return nil
					}
				}
				if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
					return nil
				}

				source, readErr := os.ReadFile(path)
				if readErr != nil {
					return fmt.Errorf("read %q: %w", path, readErr)
				}
				parsed, parseErr := parser.ParseFile(
					fileSet,
					path,
					source,
					parser.SkipObjectResolution,
				)
				if parseErr != nil {
					return fmt.Errorf("parse %q: %w", path, parseErr)
				}
				relativePath, relativeErr := filepath.Rel(absoluteRoot, path)
				if relativeErr != nil {
					return fmt.Errorf("resolve relative path for %q: %w", path, relativeErr)
				}

				report.FilesScanned++
				analyzeParsedFile(
					fileSet,
					parsed,
					filepath.ToSlash(relativePath),
					&report,
				)
				return nil
			},
		)
		if walkErr != nil {
			return auditReport{}, walkErr
		}
	}

	sort.Slice(report.Findings, func(left, right int) bool {
		if report.Findings[left].Path != report.Findings[right].Path {
			return report.Findings[left].Path < report.Findings[right].Path
		}
		if report.Findings[left].Line != report.Findings[right].Line {
			return report.Findings[left].Line < report.Findings[right].Line
		}
		if report.Findings[left].Parameter != report.Findings[right].Parameter {
			return report.Findings[left].Parameter < report.Findings[right].Parameter
		}
		return report.Findings[left].Root < report.Findings[right].Root
	})

	return report, nil
}

func analyzeParsedFile(
	fileSet *token.FileSet,
	file *ast.File,
	path string,
	report *auditReport,
) {
	aliases, dotImport := contextImportAliases(file)
	if len(aliases) == 0 && !dotImport {
		return
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.FuncDecl:
			analyzeFunction(
				fileSet,
				typed.Type,
				typed.Body,
				path,
				aliases,
				dotImport,
				report,
			)
		case *ast.FuncLit:
			analyzeFunction(
				fileSet,
				typed.Type,
				typed.Body,
				path,
				aliases,
				dotImport,
				report,
			)
		}
		return true
	})
}

func analyzeFunction(
	fileSet *token.FileSet,
	functionType *ast.FuncType,
	body *ast.BlockStmt,
	path string,
	aliases map[string]struct{},
	dotImport bool,
	report *auditReport,
) {
	if functionType == nil || functionType.Params == nil || body == nil {
		return
	}

	parameters := make(map[string]struct{})
	for _, field := range functionType.Params.List {
		if !isContextType(field.Type, aliases, dotImport) {
			continue
		}
		if len(field.Names) == 0 {
			report.ContextParameters++
			continue
		}
		for _, name := range field.Names {
			report.ContextParameters++
			parameters[name.Name] = struct{}{}
		}
	}
	if len(parameters) == 0 {
		return
	}

	firstNode := true
	ast.Inspect(body, func(node ast.Node) bool {
		if firstNode {
			firstNode = false
			return true
		}
		if _, nested := node.(*ast.FuncLit); nested {
			return false
		}

		assignment, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for index, left := range assignment.Lhs {
			identifier, ok := left.(*ast.Ident)
			if !ok {
				continue
			}
			if _, exists := parameters[identifier.Name]; !exists {
				continue
			}
			right, exists := replacementExpression(assignment, index)
			if !exists {
				continue
			}
			root, found := contextRootCall(right, aliases, dotImport)
			if !found {
				continue
			}
			report.Findings = append(report.Findings, finding{
				Path:      path,
				Line:      fileSet.Position(assignment.Pos()).Line,
				Parameter: identifier.Name,
				Root:      root,
			})
		}
		return true
	})
}

func replacementExpression(
	assignment *ast.AssignStmt,
	leftIndex int,
) (ast.Expr, bool) {
	if assignment == nil ||
		leftIndex < 0 ||
		leftIndex >= len(assignment.Lhs) ||
		len(assignment.Rhs) == 0 {
		return nil, false
	}
	if len(assignment.Lhs) == len(assignment.Rhs) {
		return assignment.Rhs[leftIndex], true
	}
	if len(assignment.Rhs) == 1 {
		return assignment.Rhs[0], true
	}
	return nil, false
}

func contextImportAliases(file *ast.File) (map[string]struct{}, bool) {
	aliases := make(map[string]struct{})
	dotImport := false

	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil || path != "context" {
			continue
		}
		if imported.Name == nil {
			aliases["context"] = struct{}{}
			continue
		}
		switch imported.Name.Name {
		case ".":
			dotImport = true
		case "_":
			continue
		default:
			aliases[imported.Name.Name] = struct{}{}
		}
	}
	return aliases, dotImport
}

func isContextType(
	expression ast.Expr,
	aliases map[string]struct{},
	dotImport bool,
) bool {
	switch typed := expression.(type) {
	case *ast.SelectorExpr:
		identifier, ok := typed.X.(*ast.Ident)
		if !ok || typed.Sel.Name != "Context" {
			return false
		}
		_, exists := aliases[identifier.Name]
		return exists
	case *ast.Ident:
		return dotImport && typed.Name == "Context"
	case *ast.ParenExpr:
		return isContextType(typed.X, aliases, dotImport)
	default:
		return false
	}
}

func contextRootCall(
	expression ast.Expr,
	aliases map[string]struct{},
	dotImport bool,
) (string, bool) {
	root := ""
	ast.Inspect(expression, func(node ast.Node) bool {
		if root != "" {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		switch function := call.Fun.(type) {
		case *ast.SelectorExpr:
			identifier, ok := function.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, exists := aliases[identifier.Name]; !exists {
				return true
			}
			if function.Sel.Name == "Background" || function.Sel.Name == "TODO" {
				root = function.Sel.Name
				return false
			}
		case *ast.Ident:
			if dotImport && (function.Name == "Background" || function.Name == "TODO") {
				root = function.Name
				return false
			}
		}
		return true
	})
	return root, root != ""
}
