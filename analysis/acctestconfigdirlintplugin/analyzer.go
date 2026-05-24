// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package acctestconfigdirlint

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	resourcePkg  = "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acctestPkg   = "github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	namedDirFunc = "NamedTestCaseDirectory"
)

var Analyzer = &analysis.Analyzer{
	Name: "acctestconfigdirlint",
	Doc:  "enforce directory-backed fixtures and step-local provider wiring in acceptance tests (resource.TestCase must be a composite literal as the second argument to resource.Test or resource.ParallelTest)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	fileLineCache := make(map[string][]string)
	varSpecIndex := buildVarSpecIndex(pass)

	for _, file := range pass.Files {
		// Skip non-test files early; varSpecIndex already used TypesInfo once up-front.
		filename := pass.Fset.File(file.Pos()).Name()
		if !strings.HasSuffix(filename, "_test.go") {
			continue
		}

		aliases := buildResourceImportAliases(file)
		if len(aliases) == 0 {
			continue
		}

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Name == nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if !isCandidateCallExpr(call) {
					return true
				}
				if !isAcceptanceTestCall(call, aliases) {
					return true
				}
				if len(call.Args) < 2 {
					return true
				}
				inspectTestCase(pass, fileLineCache, varSpecIndex, call.Args[1])
				return true
			})
		}
	}

	return nil, nil
}

// buildResourceImportAliases returns the set of local names that refer to the
// terraform-plugin-testing helper/resource package in file. Returns nil when
// the package is not imported.
func buildResourceImportAliases(file *ast.File) map[string]bool {
	aliases := make(map[string]bool)
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || path != resourcePkg {
			continue
		}
		if imp.Name != nil {
			switch imp.Name.Name {
			case "_":
				continue
			case ".":
				// Dot-import unqualified calls are out of scope; record sentinel only.
				aliases["."] = true
			default:
				aliases[imp.Name.Name] = true
			}
			continue
		}
		aliases["resource"] = true
	}
	if len(aliases) == 0 {
		return nil
	}
	return aliases
}

// isCandidateCallExpr returns true if call is syntactically a selector call whose
// selector name is "Test" or "ParallelTest". This is a cheap guard before import-alias matching.
func isCandidateCallExpr(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	return name == "Test" || name == "ParallelTest"
}

// isAcceptanceTestCall returns true if call is resource.Test(...) or resource.ParallelTest(...)
// based on syntactic import aliases (no TypesInfo lookup).
func isAcceptanceTestCall(call *ast.CallExpr, aliases map[string]bool) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if !aliases[ident.Name] {
		return false
	}
	name := sel.Sel.Name
	return name == "Test" || name == "ParallelTest"
}

// inspectTestCase extracts the Steps slice from a resource.TestCase and inspects each step.
// Only a composite literal is analyzed: patterns such as resource.Test(t, factory()) are not
// followed, so acceptance tests should pass the TestCase literal directly as the second argument.
func inspectTestCase(pass *analysis.Pass, fileLineCache map[string][]string, varSpecIndex map[*types.Var]*ast.ValueSpec, expr ast.Expr) {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return
	}

	// Confirm type is resource.TestCase.
	if !isTestCaseLit(pass, lit) {
		return
	}

	// Single pass over elements: check ProtoV6ProviderFactories and gather the Steps slice.
	var stepsLit *ast.CompositeLit
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "ProtoV6ProviderFactories":
			pass.Reportf(kv.Value.Pos(), msgTestCaseProtoV6ProviderFactories)
		case "Steps":
			if sl, ok := kv.Value.(*ast.CompositeLit); ok {
				stepsLit = sl
			}
		}
	}

	if stepsLit == nil {
		return
	}

	for _, stepElt := range stepsLit.Elts {
		stepLit, ok := stepElt.(*ast.CompositeLit)
		if !ok {
			continue
		}
		inspectTestStep(pass, fileLineCache, varSpecIndex, stepLit)
	}
}

// isTestCaseLit returns true if the composite literal is of type resource.TestCase.
func isTestCaseLit(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	if lit.Type == nil {
		// Unkeyed or inferred type – check via type info.
		t := pass.TypesInfo.TypeOf(lit)
		return isNamedType(t, resourcePkg, "TestCase")
	}

	t := pass.TypesInfo.TypeOf(lit.Type)
	return isNamedType(t, resourcePkg, "TestCase")
}

// inspectTestStep inspects a single resource.TestStep composite literal for violations.
func inspectTestStep(pass *analysis.Pass, fileLineCache map[string][]string, varSpecIndex map[*types.Var]*ast.ValueSpec, lit *ast.CompositeLit) {
	// Confirm it's actually a resource.TestStep.
	t := pass.TypesInfo.TypeOf(lit)
	if !isNamedType(t, resourcePkg, "TestStep") {
		// Try via lit.Type field.
		if lit.Type != nil {
			t = pass.TypesInfo.TypeOf(lit.Type)
		}
		if !isNamedType(t, resourcePkg, "TestStep") {
			return
		}
	}

	var configExpr ast.Expr
	var configDirExpr ast.Expr
	var externalProvidersExpr ast.Expr
	var protoV6Expr ast.Expr

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "Config":
			configExpr = kv.Value
		case "ConfigDirectory":
			configDirExpr = kv.Value
		case "ExternalProviders":
			externalProvidersExpr = kv.Value
		case "ProtoV6ProviderFactories":
			protoV6Expr = kv.Value
		}
	}

	hasConfig := configExpr != nil
	hasConfigDir := configDirExpr != nil
	hasExternalProviders := externalProvidersExpr != nil
	hasProtoV6 := protoV6Expr != nil

	// Config / ConfigDirectory / ExternalProviders relationships (field-relationship rules).
	if hasConfig || hasConfigDir || hasExternalProviders {
		switch {
		case hasExternalProviders && hasConfigDir:
			// Invalid mix: ExternalProviders + ConfigDirectory.
			pass.Reportf(externalProvidersExpr.Pos(), msgExternalProvidersWithConfigDirectory)

		case hasExternalProviders && !hasConfig:
			// ExternalProviders set but no inline Config.
			pass.Reportf(externalProvidersExpr.Pos(), msgExternalProvidersWithoutConfig)

		case hasConfig && !hasExternalProviders:
			// Inline Config without ExternalProviders is invalid for ordinary steps, even if
			// ConfigDirectory is also set (must diagnose Config, not only ConfigDirectory).
			pass.Reportf(configExpr.Pos(), msgInlineConfigWithoutExternalProviders)
			// Do not also report missing provider wiring for the same step; the inline-Config
			// diagnostic is the actionable fix for ordinary coverage.
			return

		case hasConfigDir:
			// ConfigDirectory set – must be a direct call to acctest.NamedTestCaseDirectory(...).
			if !isNamedTestCaseDirectoryCall(pass, configDirExpr) {
				pass.Reportf(configDirExpr.Pos(), msgConfigDirectoryNotNamedHelper)
			}
		}
	}

	// Provider wiring: exactly one of ProtoV6ProviderFactories or ExternalProviders per step.
	if hasProtoV6 && hasExternalProviders {
		pass.Reportf(protoV6Expr.Pos(), msgMixedStepProviderWiring)
		return
	}

	// ExternalProviders compatibility steps must source Config from an embedded testdata/.../*.tf fixture.
	if hasExternalProviders && hasConfig && !hasConfigDir {
		if !isValidEmbeddedCompatConfig(pass, fileLineCache, varSpecIndex, configExpr) {
			pass.Reportf(configExpr.Pos(), msgCompatibilityConfigMustBeEmbeddedTF)
		}
	}

	if !hasProtoV6 && !hasExternalProviders {
		pass.Reportf(lit.Lbrace, msgMissingStepProviderWiring)
	}
}

// isNamedTestCaseDirectoryCall returns true if expr is a direct call to
// acctest.NamedTestCaseDirectory(...) from the acctest package.
func isNamedTestCaseDirectoryCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := calledFunction(pass, call)
	if fn == nil || fn.Pkg() == nil {
		return false
	}
	return fn.Pkg().Path() == acctestPkg && fn.Name() == namedDirFunc
}

// calledFunction resolves a call expression to the called *types.Func, if possible.
func calledFunction(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		obj, _ := pass.TypesInfo.Uses[fun].(*types.Func)
		return obj
	case *ast.SelectorExpr:
		obj, _ := pass.TypesInfo.Uses[fun.Sel].(*types.Func)
		return obj
	default:
		return nil
	}
}

// isNamedType returns true if t (or its pointer base) is a named type with the given package path and name.
func isNamedType(t types.Type, pkgPath, typeName string) bool {
	if t == nil {
		return false
	}
	// Unwrap pointer.
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Name() == typeName && obj.Pkg() != nil && obj.Pkg().Path() == pkgPath
}
