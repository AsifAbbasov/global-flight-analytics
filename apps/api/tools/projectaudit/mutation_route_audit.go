package main

import (
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

const mutationAuthorizationIdentifier = "mutationAuthorization"

var mutationRouteMethods = map[string]struct{}{
	"Post":   {},
	"Put":    {},
	"Patch":  {},
	"Delete": {},
}

type mutationRouteRegistration struct {
	FilePath  string
	Method    string
	Path      string
	Line      int
	Protected bool
}

func auditMutationRouteProtection(
	repositoryRoot string,
	output io.Writer,
) error {
	serverDirectory := filepath.Join(
		repositoryRoot,
		"apps",
		"api",
		"internal",
		"server",
	)

	registrations, err :=
		findMutationRouteRegistrations(
			serverDirectory,
		)
	if err != nil {
		return err
	}
	if len(registrations) == 0 {
		return fmt.Errorf(
			"mutation route authorization audit found no mutation routes",
		)
	}

	failures := make(
		[]string,
		0,
	)
	for _, registration := range registrations {
		if registration.Protected {
			continue
		}
		failures = append(
			failures,
			fmt.Sprintf(
				"%s:%d %s %s must use %s as the first route middleware",
				registration.FilePath,
				registration.Line,
				strings.ToUpper(
					registration.Method,
				),
				registration.Path,
				mutationAuthorizationIdentifier,
			),
		)
	}
	if len(failures) > 0 {
		sort.Strings(failures)
		return fmt.Errorf(
			"mutation route authorization failures:\n%s",
			strings.Join(
				failures,
				"\n",
			),
		)
	}

	if err := auditFrontendMutationSecretSeparation(
		repositoryRoot,
	); err != nil {
		return err
	}

	fmt.Fprintf(
		output,
		"Mutation route authorization audit: PASS (protected_routes=%d)\n",
		len(registrations),
	)
	return nil
}

func findMutationRouteRegistrations(
	serverDirectory string,
) (
	[]mutationRouteRegistration,
	error,
) {
	registrations := make(
		[]mutationRouteRegistration,
		0,
	)

	err := filepath.WalkDir(
		serverDirectory,
		func(
			path string,
			entry fs.DirEntry,
			walkErr error,
		) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() ||
				!strings.HasSuffix(
					entry.Name(),
					".go",
				) ||
				strings.HasSuffix(
					entry.Name(),
					"_test.go",
				) {
				return nil
			}

			fileSet := token.NewFileSet()
			parsed, parseErr := parser.ParseFile(
				fileSet,
				path,
				nil,
				parser.AllErrors,
			)
			if parseErr != nil {
				return fmt.Errorf(
					"parse server route source %s: %w",
					path,
					parseErr,
				)
			}

			ast.Inspect(
				parsed,
				func(node ast.Node) bool {
					call, ok :=
						node.(*ast.CallExpr)
					if !ok {
						return true
					}
					selector, ok :=
						call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					if _, mutation :=
						mutationRouteMethods[selector.Sel.Name]; !mutation {
						return true
					}

					routePath :=
						"<dynamic-path>"
					if len(call.Args) > 0 {
						if literal, literalOK :=
							call.Args[0].(*ast.BasicLit); literalOK &&
							literal.Kind ==
								token.STRING {
							if unquoted, unquoteErr :=
								strconv.Unquote(
									literal.Value,
								); unquoteErr == nil {
								routePath =
									unquoted
							}
						}
					}

					protected := false
					if len(call.Args) >= 3 {
						if identifier, identifierOK :=
							call.Args[1].(*ast.Ident); identifierOK &&
							identifier.Name ==
								mutationAuthorizationIdentifier {
							protected = true
						}
					}

					position := fileSet.Position(
						call.Pos(),
					)
					registrations = append(
						registrations,
						mutationRouteRegistration{
							FilePath: filepath.ToSlash(
								path,
							),
							Method:    selector.Sel.Name,
							Path:      routePath,
							Line:      position.Line,
							Protected: protected,
						},
					)
					return true
				},
			)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"walk server route sources: %w",
			err,
		)
	}

	sort.Slice(
		registrations,
		func(left, right int) bool {
			if registrations[left].FilePath !=
				registrations[right].FilePath {
				return registrations[left].FilePath <
					registrations[right].FilePath
			}
			return registrations[left].Line <
				registrations[right].Line
		},
	)

	return registrations, nil
}

func auditFrontendMutationSecretSeparation(
	repositoryRoot string,
) error {
	webDirectory := filepath.Join(
		repositoryRoot,
		"apps",
		"web",
	)
	forbidden := []string{
		"X-Internal-API-Key",
		"API_MUTATION_KEY_SHA256",
	}

	failures := make(
		[]string,
		0,
	)
	err := filepath.WalkDir(
		webDirectory,
		func(
			path string,
			entry fs.DirEntry,
			walkErr error,
		) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				switch entry.Name() {
				case "node_modules",
					".next":
					return filepath.SkipDir
				default:
					return nil
				}
			}

			extension := strings.ToLower(
				filepath.Ext(path),
			)
			switch extension {
			case ".ts",
				".tsx",
				".js",
				".jsx",
				".json":
			default:
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			for _, value := range forbidden {
				if strings.Contains(
					string(content),
					value,
				) {
					failures = append(
						failures,
						fmt.Sprintf(
							"%s contains forbidden mutation credential identifier %q",
							filepath.ToSlash(path),
							value,
						),
					)
				}
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf(
			"audit frontend mutation secret separation: %w",
			err,
		)
	}
	if len(failures) > 0 {
		sort.Strings(failures)
		return fmt.Errorf(
			"frontend mutation secret separation failures:\n%s",
			strings.Join(
				failures,
				"\n",
			),
		)
	}

	return nil
}
