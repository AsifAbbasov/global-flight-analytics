package contextaudit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Policy struct {
	GuardFunction               string
	RequiredGuardedFunctions    []string
	IndependentContextFunctions map[string]string
}

type Violation struct {
	Path     string
	Line     int
	Function string
	Message  string
}

func (violation Violation) String() string {
	location := violation.Path
	if violation.Line > 0 {
		location = fmt.Sprintf("%s:%d", location, violation.Line)
	}
	if violation.Function == "" {
		return fmt.Sprintf("%s: %s", location, violation.Message)
	}
	return fmt.Sprintf("%s: function %s: %s", location, violation.Function, violation.Message)
}

func MigratorPolicy() Policy {
	return Policy{
		GuardFunction: "requireMigrationContext",
		RequiredGuardedFunctions: []string{
			"EnsureSchemaMigrations",
			"ensureSchemaMigrations",
			"Status",
			"ApplyPending",
			"applyMigrationAtomically",
			"withMigrationLock",
			"appliedMigrations",
			"appliedMigrationsWith",
		},
		IndependentContextFunctions: map[string]string{
			"releaseMigrationLock":         "migrationLockReleaseTimeout",
			"destroyLockedConnection":      "migrationLockReleaseTimeout",
			"rollbackMigrationTransaction": "migrationLockReleaseTimeout",
		},
	}
}

type contextBinding struct {
	name string
	dot  bool
}

type functionState struct {
	name                        string
	contextParameters           map[string]struct{}
	guardCallFound              bool
	contextParameterAssignments int
	withTimeoutCalls            int
	validIndependentSources     int
}

func AuditDirectory(directory string, policy Policy) ([]Violation, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read context policy directory: %w", err)
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sort.Strings(fileNames)

	fset := token.NewFileSet()
	observedGuarded := make(map[string]bool, len(policy.RequiredGuardedFunctions))
	observedIndependent := make(map[string]bool, len(policy.IndependentContextFunctions))
	violations := make([]Violation, 0)

	for _, fileName := range fileNames {
		path := filepath.Join(directory, fileName)
		file, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, parseErr)
		}

		binding, imported := findContextBinding(file)
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.FuncDecl:
				state := auditFunction(
					fset,
					directory,
					typed,
					binding,
					imported,
					policy,
					&violations,
				)
				if contains(policy.RequiredGuardedFunctions, state.name) {
					observedGuarded[state.name] = true
					if len(state.contextParameters) == 0 {
						violations = append(violations, newViolation(
							fset,
							directory,
							typed.Pos(),
							state.name,
							"required database-reaching boundary has no context.Context parameter",
						))
					}
					if !state.guardCallFound {
						violations = append(violations, newViolation(
							fset,
							directory,
							typed.Pos(),
							state.name,
							fmt.Sprintf("must call %s with its caller-owned context parameter", policy.GuardFunction),
						))
					}
				}

				if _, allowed := policy.IndependentContextFunctions[state.name]; allowed {
					observedIndependent[state.name] = true
					if state.withTimeoutCalls != 1 || state.validIndependentSources != 1 {
						violations = append(violations, newViolation(
							fset,
							directory,
							typed.Pos(),
							state.name,
							"independent cleanup context must use exactly one context.WithTimeout(context.Background(), configuredTimeout) expression",
						))
					}
				}
			case *ast.GenDecl:
				auditNode(
					fset,
					directory,
					typed,
					"",
					binding,
					imported,
					policy,
					nil,
					&violations,
				)
			}
		}
	}

	for _, functionName := range policy.RequiredGuardedFunctions {
		if !observedGuarded[functionName] {
			violations = append(violations, Violation{
				Path:     filepath.Base(directory),
				Function: functionName,
				Message:  "required guarded function is missing",
			})
		}
	}
	for functionName := range policy.IndependentContextFunctions {
		if !observedIndependent[functionName] {
			violations = append(violations, Violation{
				Path:     filepath.Base(directory),
				Function: functionName,
				Message:  "required independent cleanup function is missing",
			})
		}
	}

	sort.Slice(violations, func(left int, right int) bool {
		if violations[left].Path != violations[right].Path {
			return violations[left].Path < violations[right].Path
		}
		if violations[left].Line != violations[right].Line {
			return violations[left].Line < violations[right].Line
		}
		if violations[left].Function != violations[right].Function {
			return violations[left].Function < violations[right].Function
		}
		return violations[left].Message < violations[right].Message
	})
	return violations, nil
}

func auditFunction(
	fset *token.FileSet,
	directory string,
	function *ast.FuncDecl,
	binding contextBinding,
	contextImported bool,
	policy Policy,
	violations *[]Violation,
) functionState {
	state := functionState{
		name:              function.Name.Name,
		contextParameters: contextParameterNames(function, binding, contextImported),
	}
	if function.Body == nil {
		return state
	}

	auditNode(
		fset,
		directory,
		function.Body,
		state.name,
		binding,
		contextImported,
		policy,
		&state,
		violations,
	)
	return state
}

func auditNode(
	fset *token.FileSet,
	directory string,
	root ast.Node,
	functionName string,
	binding contextBinding,
	contextImported bool,
	policy Policy,
	state *functionState,
	violations *[]Violation,
) {
	if root == nil || !contextImported {
		return
	}

	var stack []ast.Node
	ast.Inspect(root, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return false
		}

		var parent ast.Node
		if len(stack) > 0 {
			parent = stack[len(stack)-1]
		}
		stack = append(stack, node)

		switch typed := node.(type) {
		case *ast.AssignStmt:
			if state != nil {
				for _, left := range typed.Lhs {
					identifier, ok := left.(*ast.Ident)
					if !ok {
						continue
					}
					if _, isContextParameter := state.contextParameters[identifier.Name]; isContextParameter {
						state.contextParameterAssignments++
						*violations = append(*violations, newViolation(
							fset,
							directory,
							identifier.Pos(),
							functionName,
							fmt.Sprintf("caller-owned context parameter %s must not be reassigned", identifier.Name),
						))
					}
				}
			}
		case *ast.CallExpr:
			member, ok := contextMember(typed.Fun, binding)
			if ok {
				auditContextCall(
					fset,
					directory,
					typed,
					parent,
					functionName,
					member,
					binding,
					policy,
					state,
					violations,
				)
			}
			if state != nil && isGuardCall(typed, policy.GuardFunction, state.contextParameters) {
				state.guardCallFound = true
			}
		case *ast.SelectorExpr:
			if call, ok := parent.(*ast.CallExpr); ok && call.Fun == typed {
				break
			}
			member, ok := contextMember(typed, binding)
			if ok && isForbiddenSourceMember(member) {
				*violations = append(*violations, newViolation(
					fset,
					directory,
					typed.Pos(),
					functionName,
					fmt.Sprintf("context.%s must not be stored or passed as a function value", member),
				))
			}
		case *ast.Ident:
			if !binding.dot {
				break
			}
			if call, ok := parent.(*ast.CallExpr); ok && call.Fun == typed {
				break
			}
			if isForbiddenSourceMember(typed.Name) {
				*violations = append(*violations, newViolation(
					fset,
					directory,
					typed.Pos(),
					functionName,
					fmt.Sprintf("context.%s must not be stored or passed as a function value", typed.Name),
				))
			}
		}

		return true
	})
}

func auditContextCall(
	fset *token.FileSet,
	directory string,
	call *ast.CallExpr,
	parent ast.Node,
	functionName string,
	member string,
	binding contextBinding,
	policy Policy,
	state *functionState,
	violations *[]Violation,
) {
	timeoutIdentifier, independentFunction := policy.IndependentContextFunctions[functionName]

	switch member {
	case "Background":
		if independentFunction && expectedIndependentBackground(call, parent, binding, timeoutIdentifier) {
			if state != nil {
				state.validIndependentSources++
			}
			return
		}
		*violations = append(*violations, newViolation(
			fset,
			directory,
			call.Pos(),
			functionName,
			"context.Background() is forbidden outside the exact bounded cleanup pattern",
		))
	case "TODO":
		*violations = append(*violations, newViolation(
			fset,
			directory,
			call.Pos(),
			functionName,
			"context.TODO() is forbidden in production migrator code",
		))
	case "WithoutCancel":
		*violations = append(*violations, newViolation(
			fset,
			directory,
			call.Pos(),
			functionName,
			"context.WithoutCancel() must not detach database-reaching work from caller cancellation",
		))
	case "WithTimeout":
		if !independentFunction {
			return
		}
		if state != nil {
			state.withTimeoutCalls++
		}
		if !expectedIndependentWithTimeout(call, binding, timeoutIdentifier) {
			*violations = append(*violations, newViolation(
				fset,
				directory,
				call.Pos(),
				functionName,
				fmt.Sprintf("cleanup context must be context.WithTimeout(context.Background(), %s)", timeoutIdentifier),
			))
		}
	}
}

func expectedIndependentBackground(
	backgroundCall *ast.CallExpr,
	parent ast.Node,
	binding contextBinding,
	timeoutIdentifier string,
) bool {
	withTimeout, ok := parent.(*ast.CallExpr)
	if !ok || len(withTimeout.Args) != 2 || withTimeout.Args[0] != backgroundCall {
		return false
	}
	member, ok := contextMember(withTimeout.Fun, binding)
	if !ok || member != "WithTimeout" {
		return false
	}
	identifier, ok := withTimeout.Args[1].(*ast.Ident)
	return ok && identifier.Name == timeoutIdentifier
}

func expectedIndependentWithTimeout(
	call *ast.CallExpr,
	binding contextBinding,
	timeoutIdentifier string,
) bool {
	if len(call.Args) != 2 {
		return false
	}
	backgroundCall, ok := call.Args[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	member, ok := contextMember(backgroundCall.Fun, binding)
	if !ok || member != "Background" || len(backgroundCall.Args) != 0 {
		return false
	}
	identifier, ok := call.Args[1].(*ast.Ident)
	return ok && identifier.Name == timeoutIdentifier
}

func contextParameterNames(
	function *ast.FuncDecl,
	binding contextBinding,
	contextImported bool,
) map[string]struct{} {
	names := make(map[string]struct{})
	if !contextImported || function.Type == nil || function.Type.Params == nil {
		return names
	}
	for _, field := range function.Type.Params.List {
		if !isContextType(field.Type, binding) {
			continue
		}
		for _, name := range field.Names {
			names[name.Name] = struct{}{}
		}
	}
	return names
}

func isContextType(expression ast.Expr, binding contextBinding) bool {
	member, ok := contextMember(expression, binding)
	return ok && member == "Context"
}

func isGuardCall(call *ast.CallExpr, guardFunction string, contextParameters map[string]struct{}) bool {
	identifier, ok := call.Fun.(*ast.Ident)
	if !ok || identifier.Name != guardFunction || len(call.Args) != 1 {
		return false
	}
	argument, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return false
	}
	_, ok = contextParameters[argument.Name]
	return ok
}

func findContextBinding(file *ast.File) (contextBinding, bool) {
	for _, imported := range file.Imports {
		path, err := strconv.Unquote(imported.Path.Value)
		if err != nil || path != "context" {
			continue
		}
		if imported.Name == nil {
			return contextBinding{name: "context"}, true
		}
		switch imported.Name.Name {
		case "_":
			return contextBinding{}, false
		case ".":
			return contextBinding{dot: true}, true
		default:
			return contextBinding{name: imported.Name.Name}, true
		}
	}
	return contextBinding{}, false
}

func contextMember(expression ast.Expr, binding contextBinding) (string, bool) {
	if binding.dot {
		identifier, ok := expression.(*ast.Ident)
		if !ok {
			return "", false
		}
		return identifier.Name, true
	}
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	packageIdentifier, ok := selector.X.(*ast.Ident)
	if !ok || packageIdentifier.Name != binding.name {
		return "", false
	}
	return selector.Sel.Name, true
}

func isForbiddenSourceMember(member string) bool {
	switch member {
	case "Background", "TODO", "WithoutCancel":
		return true
	default:
		return false
	}
}

func newViolation(
	fset *token.FileSet,
	directory string,
	position token.Pos,
	functionName string,
	message string,
) Violation {
	resolved := fset.Position(position)
	path := resolved.Filename
	if relative, err := filepath.Rel(directory, path); err == nil {
		path = filepath.ToSlash(relative)
	}
	return Violation{
		Path:     path,
		Line:     resolved.Line,
		Function: functionName,
		Message:  message,
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
