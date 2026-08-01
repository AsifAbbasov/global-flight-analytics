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

const auditVersion = "backend-timeout-consistency-audit-v1"

type finding struct {
	Path   string
	Line   int
	Rule   string
	Detail string
}

type auditReport struct {
	FilesScanned              int
	HTTPFiles                 int
	ContextualRequests        int
	BoundedHTTPClientLiterals int
	HandlerUserContexts       int
	Findings                  []finding
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args[1:]))
}

func run(stdout io.Writer, stderr io.Writer, args []string) int {
	flags := flag.NewFlagSet("backendtimeoutconsistencyaudit", flag.ContinueOnError)
	flags.SetOutput(stderr)

	root := flags.String("root", ".", "Go module root to audit")
	strict := flags.Bool("strict", false, "fail when findings exist")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	report, err := scanRepository(*root)
	if err != nil {
		fmt.Fprintf(stderr, "backend timeout consistency audit: %v\n", err)
		return 2
	}

	fmt.Fprintf(stdout, "Audit version: %s\n", auditVersion)
	fmt.Fprintf(stdout, "Production Go files scanned: %d\n", report.FilesScanned)
	fmt.Fprintf(stdout, "Files importing net/http: %d\n", report.HTTPFiles)
	fmt.Fprintf(stdout, "Context-bound HTTP requests: %d\n", report.ContextualRequests)
	fmt.Fprintf(stdout, "Bounded HTTP client literals: %d\n", report.BoundedHTTPClientLiterals)
	fmt.Fprintf(stdout, "Handler user-context calls: %d\n", report.HandlerUserContexts)

	if len(report.Findings) == 0 {
		fmt.Fprintln(stdout, "Timeout consistency findings: 0")
		fmt.Fprintln(stdout, "Backend timeout consistency audit: PASS")
		return 0
	}

	fmt.Fprintf(stdout, "Timeout consistency findings: %d\n", len(report.Findings))
	fmt.Fprintln(stdout, "Backend timeout consistency audit: FAIL")
	for _, item := range report.Findings {
		fmt.Fprintf(
			stdout,
			"- %s:%d [%s] %s\n",
			item.Path,
			item.Line,
			item.Rule,
			item.Detail,
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
		if report.Findings[left].Rule != report.Findings[right].Rule {
			return report.Findings[left].Rule < report.Findings[right].Rule
		}
		return report.Findings[left].Detail < report.Findings[right].Detail
	})

	return report, nil
}

func analyzeParsedFile(
	fileSet *token.FileSet,
	file *ast.File,
	path string,
	report *auditReport,
) {
	httpAliases, dotImport := importedPackageAliases(file, "net/http", "http")
	if dotImport {
		addFinding(
			fileSet,
			file,
			path,
			"http-dot-import",
			"dot imports of net/http make timeout ownership impossible to audit",
			report,
		)
	}
	if len(httpAliases) > 0 || dotImport {
		report.HTTPFiles++
	}

	handlerFile := strings.HasPrefix(path, "internal/http/handlers/")

	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.CallExpr:
			selector, ok := typed.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if handlerFile {
				switch selector.Sel.Name {
				case "Context":
					addFinding(
						fileSet,
						typed,
						path,
						"fiber-transport-context",
						"HTTP handlers must pass c.UserContext() so request deadlines reach services and repositories",
						report,
					)
				case "UserContext":
					report.HandlerUserContexts++
				}
			}

			packageName, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, imported := httpAliases[packageName.Name]; !imported {
				return true
			}

			switch selector.Sel.Name {
			case "NewRequestWithContext":
				report.ContextualRequests++
			case "NewRequest":
				addFinding(
					fileSet,
					typed,
					path,
					"http-request-without-context",
					"use http.NewRequestWithContext with a caller-owned bounded context",
					report,
				)
			case "Get", "Head", "Post", "PostForm":
				addFinding(
					fileSet,
					typed,
					path,
					"http-package-shortcut",
					"package-level HTTP shortcuts use the default client and do not expose an explicit timeout contract",
					report,
				)
			}

		case *ast.SelectorExpr:
			packageName, ok := typed.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, imported := httpAliases[packageName.Name]; !imported {
				return true
			}
			if typed.Sel.Name == "DefaultClient" || typed.Sel.Name == "DefaultTransport" {
				addFinding(
					fileSet,
					typed,
					path,
					"http-default-global",
					"default HTTP globals hide timeout and transport ownership",
					report,
				)
			}

		case *ast.CompositeLit:
			selector, ok := typed.Type.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Client" {
				return true
			}
			packageName, ok := selector.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, imported := httpAliases[packageName.Name]; !imported {
				return true
			}

			timeoutExpression, found := clientTimeoutExpression(typed)
			if !found {
				addFinding(
					fileSet,
					typed,
					path,
					"http-client-timeout-missing",
					"http.Client literals must declare Timeout explicitly",
					report,
				)
				return true
			}
			if isLiteralNonPositive(timeoutExpression) {
				addFinding(
					fileSet,
					timeoutExpression,
					path,
					"http-client-timeout-nonpositive",
					"http.Client Timeout must be greater than zero",
					report,
				)
				return true
			}
			report.BoundedHTTPClientLiterals++
		}

		return true
	})
}

func importedPackageAliases(
	file *ast.File,
	importPath string,
	defaultName string,
) (map[string]struct{}, bool) {
	aliases := make(map[string]struct{})
	dotImport := false

	for _, item := range file.Imports {
		path, err := strconv.Unquote(item.Path.Value)
		if err != nil || path != importPath {
			continue
		}

		if item.Name == nil {
			aliases[defaultName] = struct{}{}
			continue
		}

		switch item.Name.Name {
		case ".":
			dotImport = true
		case "_":
		default:
			aliases[item.Name.Name] = struct{}{}
		}
	}

	return aliases, dotImport
}

func clientTimeoutExpression(
	literal *ast.CompositeLit,
) (ast.Expr, bool) {
	for _, element := range literal.Elts {
		keyValue, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := keyValue.Key.(*ast.Ident)
		if !ok || key.Name != "Timeout" {
			continue
		}
		return keyValue.Value, true
	}
	return nil, false
}

func isLiteralNonPositive(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		if typed.Kind != token.INT && typed.Kind != token.FLOAT {
			return false
		}
		value, err := strconv.ParseFloat(typed.Value, 64)
		return err == nil && value <= 0
	case *ast.UnaryExpr:
		if typed.Op != token.SUB {
			return false
		}
		_, ok := typed.X.(*ast.BasicLit)
		return ok
	case *ast.CallExpr:
		if len(typed.Args) != 1 {
			return false
		}
		return isLiteralNonPositive(typed.Args[0])
	default:
		return false
	}
}

func addFinding(
	fileSet *token.FileSet,
	node ast.Node,
	path string,
	rule string,
	detail string,
	report *auditReport,
) {
	line := 1
	if node != nil {
		line = fileSet.Position(node.Pos()).Line
	}
	report.Findings = append(
		report.Findings,
		finding{
			Path:   path,
			Line:   line,
			Rule:   rule,
			Detail: detail,
		},
	)
}
