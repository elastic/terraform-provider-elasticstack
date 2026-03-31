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
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	resourcePkg  = "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acctestPkg   = "github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	namedDirFunc = "NamedTestCaseDirectory"
)

var Analyzer = &analysis.Analyzer{
	Name:     "acctestconfigdirlint",
	Doc:      "enforce directory-backed fixture usage in acceptance test steps",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Only process _test.go files.
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		// Check if this is a call to resource.Test or resource.ParallelTest.
		if !isAcceptanceTestCall(pass, call) {
			return
		}

		// The call is resource.Test(t, testCase) or resource.ParallelTest(t, testCase).
		// The second argument should be the resource.TestCase.
		if len(call.Args) < 2 {
			return
		}

		// Get the filename to check it's a _test.go file.
		pos := pass.Fset.Position(call.Pos())
		if !strings.HasSuffix(pos.Filename, "_test.go") {
			return
		}

		testCaseArg := call.Args[1]
		inspectTestCase(pass, testCaseArg)
	})

	return nil, nil
}

// isAcceptanceTestCall returns true if call is resource.Test(...) or resource.ParallelTest(...).
func isAcceptanceTestCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	fnObj := calledFunction(pass, call)
	if fnObj == nil || fnObj.Pkg() == nil {
		return false
	}
	if fnObj.Pkg().Path() != resourcePkg {
		return false
	}
	name := fnObj.Name()
	return name == "Test" || name == "ParallelTest"
}

// inspectTestCase extracts the Steps slice from a resource.TestCase and inspects each step.
func inspectTestCase(pass *analysis.Pass, expr ast.Expr) {
	// The TestCase may be a composite literal or a variable reference.
	// Handle composite literal directly; for variables, we can't easily follow them,
	// so we only handle the inline literal case.
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return
	}

	// Confirm type is resource.TestCase.
	if !isTestCaseLit(pass, lit) {
		return
	}

	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Steps" {
			continue
		}

		// The value should be a slice literal of resource.TestStep.
		sliceLit, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		for _, stepElt := range sliceLit.Elts {
			stepLit, ok := stepElt.(*ast.CompositeLit)
			if !ok {
				continue
			}
			inspectTestStep(pass, stepLit)
		}
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
func inspectTestStep(pass *analysis.Pass, lit *ast.CompositeLit) {
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
		}
	}

	hasConfig := configExpr != nil
	hasConfigDir := configDirExpr != nil
	hasExternalProviders := externalProvidersExpr != nil

	// Non-goal: steps with neither Config, ConfigDirectory, nor ExternalProviders are out of scope.
	// Import-only, refresh-only, and plan-only steps need not specify a config source.
	if !hasConfig && !hasConfigDir && !hasExternalProviders {
		return
	}

	switch {
	case hasExternalProviders && hasConfigDir:
		// Invalid mix: ExternalProviders + ConfigDirectory.
		pass.Reportf(externalProvidersExpr.Pos(), msgExternalProvidersWithConfigDirectory)

	case hasExternalProviders && !hasConfig:
		// ExternalProviders set but no inline Config.
		pass.Reportf(externalProvidersExpr.Pos(), msgExternalProvidersWithoutConfig)

	case hasConfigDir:
		// ConfigDirectory set – must be a direct call to acctest.NamedTestCaseDirectory(...).
		if !isNamedTestCaseDirectoryCall(pass, configDirExpr) {
			pass.Reportf(configDirExpr.Pos(), msgConfigDirectoryNotNamedHelper)
		}

	case hasConfig && !hasExternalProviders:
		// Ordinary step with inline Config but no ExternalProviders.
		pass.Reportf(configExpr.Pos(), msgInlineConfigWithoutExternalProviders)
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
