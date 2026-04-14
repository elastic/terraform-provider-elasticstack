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

package esclienthelper

import (
	"go/ast"
	"go/types"
	"maps"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	modulePath = "github.com/elastic/terraform-provider-elasticstack"
	clientsPkg = modulePath + "/internal/clients"
	esPkg      = modulePath + "/internal/clients/elasticsearch"
)

type Config struct {
	Wrappers []string `json:"wrappers"`
}

func NewAnalyzer(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "esclienthelper",
		Doc:  "enforce helper-derived API clients at Elasticsearch sink calls",
		FactTypes: []analysis.Fact{
			&clientReturnFact{},
		},
		Run: func(pass *analysis.Pass) (any, error) {
			allowedWrappers := toSet(cfg.Wrappers)

			// Precompute the in-scope non-test Elasticsearch files once per pass
			// and reuse across both the fact-export and sink-check phases.
			inScopeFiles := collectInScopeFiles(pass)

			// Run-scoped caches: keyed by *types.Func for stable per-function metadata.
			// These stop repeated signature rescans and fact reimports for the same function.
			sinkParamCache := make(map[*types.Func][]int)
			factCache := make(map[*types.Func]*clientReturnFact)

			// Phase 1: export facts for functions whose first result is an *APIClient.
			for _, file := range inScopeFiles {
				for _, decl := range file.Decls {
					fd, ok := decl.(*ast.FuncDecl)
					if !ok || fd.Body == nil {
						continue
					}
					exportFunctionFacts(pass, fd, allowedWrappers)
				}
			}

			// Phase 2: check sink calls in in-scope functions.
			for _, file := range inScopeFiles {
				for _, decl := range file.Decls {
					fd, ok := decl.(*ast.FuncDecl)
					if !ok || fd.Body == nil {
						continue
					}
					inspectFunction(pass, fd, allowedWrappers, sinkParamCache, factCache)
				}
			}

			return nil, nil
		},
	}
}

var Analyzer = NewAnalyzer(Config{})

// collectInScopeFiles returns the subset of pass.Files that are non-test files
// located under an internal/elasticsearch/ directory. This precomputation is
// done once per pass and shared across both analysis phases.
func collectInScopeFiles(pass *analysis.Pass) []*ast.File {
	result := make([]*ast.File, 0, len(pass.Files))
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if isInElasticsearchDir(filename) && !strings.HasSuffix(filename, "_test.go") {
			result = append(result, file)
		}
	}
	return result
}

type clientReturnFact struct {
	AlwaysDerived        bool
	RequiredParamIndexes []int
}

func (*clientReturnFact) AFact() {}

func (f *clientReturnFact) String() string {
	if f == nil {
		return "client-return: <nil>"
	}
	if f.AlwaysDerived {
		return "client-return: always-derived"
	}
	if len(f.RequiredParamIndexes) == 0 {
		return "client-return: unknown"
	}
	return "client-return: requires-params"
}

func inspectFunction(pass *analysis.Pass, fd *ast.FuncDecl, allowedWrappers map[string]struct{}, sinkParamCache map[*types.Func][]int, factCache map[*types.Func]*clientReturnFact) {
	derivedVars := map[*types.Var]bool{}
	inspectBlock(pass, fd.Body, allowedWrappers, derivedVars, sinkParamCache, factCache)
}

func exportFunctionFacts(pass *analysis.Pass, fd *ast.FuncDecl, allowedWrappers map[string]struct{}) {
	fnObj, _ := pass.TypesInfo.Defs[fd.Name].(*types.Func)
	if fnObj == nil {
		return
	}
	sig, ok := fnObj.Type().(*types.Signature)
	if !ok || sig.Results().Len() == 0 || !isAPIClientPointer(sig.Results().At(0).Type()) {
		return
	}

	paramIndexByVar := map[*types.Var]int{}
	for i := 0; i < sig.Params().Len(); i++ {
		param := sig.Params().At(i)
		if param != nil {
			paramIndexByVar[param] = i
		}
	}

	returnStmts := collectReturnStatements(fd.Body)
	if len(returnStmts) == 0 {
		return
	}

	required := map[int]struct{}{}
	for _, ret := range returnStmts {
		if len(ret.Results) == 0 {
			return
		}
		reqs, ok := inferRequiredParamsForDerived(pass, ret.Results[0], paramIndexByVar, allowedWrappers)
		if !ok {
			return
		}
		for idx := range reqs {
			required[idx] = struct{}{}
		}
	}

	requiredList := make([]int, 0, len(required))
	for idx := range required {
		requiredList = append(requiredList, idx)
	}
	sort.Ints(requiredList)

	fact := &clientReturnFact{
		AlwaysDerived:        len(requiredList) == 0,
		RequiredParamIndexes: requiredList,
	}
	pass.ExportObjectFact(fnObj, fact)
}

func collectReturnStatements(body *ast.BlockStmt) []*ast.ReturnStmt {
	results := make([]*ast.ReturnStmt, 0)
	ast.Inspect(body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		results = append(results, ret)
		return true
	})
	return results
}

func inferRequiredParamsForDerived(pass *analysis.Pass, expr ast.Expr, paramIndexByVar map[*types.Var]int, allowedWrappers map[string]struct{}) (map[int]struct{}, bool) {
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return inferRequiredParamsForDerived(pass, e.X, paramIndexByVar, allowedWrappers)
	case *ast.UnaryExpr:
		return inferRequiredParamsForDerived(pass, e.X, paramIndexByVar, allowedWrappers)
	case *ast.TypeAssertExpr:
		return inferRequiredParamsForDerived(pass, e.X, paramIndexByVar, allowedWrappers)
	case *ast.Ident:
		v, ok := pass.TypesInfo.ObjectOf(e).(*types.Var)
		if !ok || v == nil {
			return nil, false
		}
		paramIndex, ok := paramIndexByVar[v]
		if !ok {
			return nil, false
		}
		return map[int]struct{}{paramIndex: {}}, true
	case *ast.CallExpr:
		if isApprovedSourceCall(pass, e, allowedWrappers) {
			return map[int]struct{}{}, true
		}
		fnObj := calledFunction(pass, e)
		if fnObj == nil {
			return nil, false
		}
		fact := &clientReturnFact{}
		if !pass.ImportObjectFact(fnObj, fact) {
			return nil, false
		}
		if fact.AlwaysDerived {
			return map[int]struct{}{}, true
		}
		reqs := map[int]struct{}{}
		for _, calleeParamIdx := range fact.RequiredParamIndexes {
			if calleeParamIdx < 0 || calleeParamIdx >= len(e.Args) {
				return nil, false
			}
			subReqs, ok := inferRequiredParamsForDerived(pass, e.Args[calleeParamIdx], paramIndexByVar, allowedWrappers)
			if !ok {
				return nil, false
			}
			for idx := range subReqs {
				reqs[idx] = struct{}{}
			}
		}
		return reqs, true
	default:
		return nil, false
	}
}

func inspectBlock(pass *analysis.Pass, body *ast.BlockStmt, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, sinkParamCache map[*types.Func][]int, factCache map[*types.Func]*clientReturnFact) {
	if body == nil {
		return
	}
	for _, stmt := range body.List {
		inspectStmt(pass, stmt, allowedWrappers, derivedVars, sinkParamCache, factCache)
	}
}

func inspectStmt(pass *analysis.Pass, stmt ast.Stmt, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, sinkParamCache map[*types.Func][]int, factCache map[*types.Func]*clientReturnFact) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		for _, rhs := range s.Rhs {
			inspectExpr(pass, rhs, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
		applyAssignment(pass, s, allowedWrappers, derivedVars, factCache)
	case *ast.DeclStmt:
		gen, ok := s.Decl.(*ast.GenDecl)
		if !ok {
			return
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, value := range vs.Values {
				inspectExpr(pass, value, allowedWrappers, derivedVars, sinkParamCache, factCache)
			}
			applyValueSpec(pass, vs, allowedWrappers, derivedVars, factCache)
		}
	case *ast.ExprStmt:
		inspectExpr(pass, s.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.ReturnStmt:
		for _, result := range s.Results {
			inspectExpr(pass, result, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
	case *ast.IfStmt:
		if s.Init != nil {
			inspectStmt(pass, s.Init, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
		inspectExpr(pass, s.Cond, allowedWrappers, derivedVars, sinkParamCache, factCache)

		thenState := copyDerivedMap(derivedVars)
		inspectBlock(pass, s.Body, allowedWrappers, thenState, sinkParamCache, factCache)

		elseState := copyDerivedMap(derivedVars)
		if s.Else != nil {
			inspectStmt(pass, s.Else, allowedWrappers, elseState, sinkParamCache, factCache)
		}
		mergeDerivedState(derivedVars, thenState, elseState)
	case *ast.BlockStmt:
		inspectBlock(pass, s, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.ForStmt:
		if s.Init != nil {
			inspectStmt(pass, s.Init, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
		if s.Cond != nil {
			inspectExpr(pass, s.Cond, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
		inspectBlock(pass, s.Body, allowedWrappers, copyDerivedMap(derivedVars), sinkParamCache, factCache)
		if s.Post != nil {
			inspectStmt(pass, s.Post, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
	case *ast.RangeStmt:
		inspectExpr(pass, s.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectBlock(pass, s.Body, allowedWrappers, copyDerivedMap(derivedVars), sinkParamCache, factCache)
	}
}

func inspectExpr(pass *analysis.Pass, expr ast.Expr, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, sinkParamCache map[*types.Func][]int, factCache map[*types.Func]*clientReturnFact) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpr:
		checkSinkCall(pass, e, allowedWrappers, derivedVars, sinkParamCache, factCache)
		for _, arg := range e.Args {
			inspectExpr(pass, arg, allowedWrappers, derivedVars, sinkParamCache, factCache)
		}
	case *ast.BinaryExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectExpr(pass, e.Y, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.UnaryExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.ParenExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.IndexExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectExpr(pass, e.Index, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.SelectorExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.TypeAssertExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
	case *ast.SliceExpr:
		inspectExpr(pass, e.X, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectExpr(pass, e.Low, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectExpr(pass, e.High, allowedWrappers, derivedVars, sinkParamCache, factCache)
		inspectExpr(pass, e.Max, allowedWrappers, derivedVars, sinkParamCache, factCache)
	}
}

func checkSinkCall(pass *analysis.Pass, call *ast.CallExpr, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, sinkParamCache map[*types.Func][]int, factCache map[*types.Func]*clientReturnFact) {
	receiverExpr, hasReceiverSink := receiverSinkExpr(pass, call)
	if hasReceiverSink && !isDerivedClientExpr(pass, receiverExpr, allowedWrappers, derivedVars, factCache) {
		pass.Reportf(receiverExpr.Pos(), sinkDiagnosticMessage)
	}

	// Resolve callee once and reuse for both the sink param cache lookup and arg checks.
	fnObj := calledFunction(pass, call)
	argIdxs := elasticsearchClientParamIndicesCached(fnObj, sinkParamCache)
	for _, argIdx := range argIdxs {
		if argIdx >= len(call.Args) {
			continue
		}
		argExpr := call.Args[argIdx]
		if !isDerivedClientExpr(pass, argExpr, allowedWrappers, derivedVars, factCache) {
			pass.Reportf(argExpr.Pos(), sinkDiagnosticMessage)
		}
	}
}

func receiverSinkExpr(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	selection, ok := pass.TypesInfo.Selections[sel]
	if !ok || selection == nil {
		return nil, false
	}
	return sel.X, isAPIClientPointer(selection.Recv())
}

// elasticsearchClientParamIndicesCached returns the parameter indices of *APIClient parameters
// for the given function, using sinkParamCache to avoid rescanning the signature on repeat calls.
// fnObj may be nil (cache will not be populated for nil keys).
func elasticsearchClientParamIndicesCached(fnObj *types.Func, sinkParamCache map[*types.Func][]int) []int {
	if fnObj == nil || fnObj.Pkg() == nil || !strings.HasPrefix(fnObj.Pkg().Path(), esPkg) {
		return nil
	}

	if cached, ok := sinkParamCache[fnObj]; ok {
		return cached
	}

	sig, ok := fnObj.Type().(*types.Signature)
	if !ok {
		sinkParamCache[fnObj] = nil
		return nil
	}

	result := make([]int, 0)
	for i := 0; i < sig.Params().Len(); i++ {
		if isAPIClientPointer(sig.Params().At(i).Type()) {
			result = append(result, i)
		}
	}
	sinkParamCache[fnObj] = result
	return result
}

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

func applyAssignment(pass *analysis.Pass, assign *ast.AssignStmt, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, factCache map[*types.Func]*clientReturnFact) {
	if len(assign.Rhs) == 1 && len(assign.Lhs) > 1 {
		derived := isDerivedClientExpr(pass, assign.Rhs[0], allowedWrappers, derivedVars, factCache)
		for _, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name == "_" {
				continue
			}
			if v, ok := pass.TypesInfo.ObjectOf(id).(*types.Var); ok && v != nil {
				derivedVars[v] = derived
			}
		}
		return
	}

	for i := range min(len(assign.Lhs), len(assign.Rhs)) {
		id, ok := assign.Lhs[i].(*ast.Ident)
		if !ok || id.Name == "_" {
			continue
		}
		v, ok := pass.TypesInfo.ObjectOf(id).(*types.Var)
		if !ok || v == nil {
			continue
		}
		derivedVars[v] = isDerivedClientExpr(pass, assign.Rhs[i], allowedWrappers, derivedVars, factCache)
	}
}

func applyValueSpec(pass *analysis.Pass, spec *ast.ValueSpec, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, factCache map[*types.Func]*clientReturnFact) {
	if len(spec.Values) == 1 && len(spec.Names) > 1 {
		derived := isDerivedClientExpr(pass, spec.Values[0], allowedWrappers, derivedVars, factCache)
		for _, name := range spec.Names {
			if v, ok := pass.TypesInfo.ObjectOf(name).(*types.Var); ok && v != nil {
				derivedVars[v] = derived
			}
		}
		return
	}

	for i := range min(len(spec.Names), len(spec.Values)) {
		v, ok := pass.TypesInfo.ObjectOf(spec.Names[i]).(*types.Var)
		if !ok || v == nil {
			continue
		}
		derivedVars[v] = isDerivedClientExpr(pass, spec.Values[i], allowedWrappers, derivedVars, factCache)
	}
}

func isDerivedClientExpr(pass *analysis.Pass, expr ast.Expr, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, factCache map[*types.Func]*clientReturnFact) bool {
	switch e := expr.(type) {
	case *ast.ParenExpr:
		return isDerivedClientExpr(pass, e.X, allowedWrappers, derivedVars, factCache)
	case *ast.Ident:
		v, ok := pass.TypesInfo.ObjectOf(e).(*types.Var)
		if !ok || v == nil {
			return false
		}
		return derivedVars[v]
	case *ast.CallExpr:
		if isApprovedSourceCall(pass, e, allowedWrappers) {
			return true
		}
		fnObj := calledFunction(pass, e)
		return isFactDerivedCallWithFn(pass, e, fnObj, allowedWrappers, derivedVars, factCache)
	case *ast.TypeAssertExpr:
		return isDerivedClientExpr(pass, e.X, allowedWrappers, derivedVars, factCache)
	case *ast.UnaryExpr:
		return isDerivedClientExpr(pass, e.X, allowedWrappers, derivedVars, factCache)
	default:
		return false
	}
}

// isFactDerivedCallWithFn checks whether the call returns a derived client by importing the fact
// for fnObj. fnObj is passed in so the caller can reuse an already-resolved *types.Func without
// calling calledFunction again. factCache avoids repeated pass.ImportObjectFact calls for the
// same function across the pass.
func isFactDerivedCallWithFn(pass *analysis.Pass, call *ast.CallExpr, fnObj *types.Func, allowedWrappers map[string]struct{}, derivedVars map[*types.Var]bool, factCache map[*types.Func]*clientReturnFact) bool {
	if fnObj == nil {
		return false
	}

	fact, ok := factCache[fnObj]
	if !ok {
		f := &clientReturnFact{}
		if pass.ImportObjectFact(fnObj, f) {
			fact = f
		}
		factCache[fnObj] = fact // nil means "no fact available"
	}

	if fact == nil {
		return false
	}

	if fact.AlwaysDerived {
		return true
	}
	if len(fact.RequiredParamIndexes) == 0 {
		return false
	}

	for _, idx := range fact.RequiredParamIndexes {
		if idx < 0 || idx >= len(call.Args) {
			return false
		}
		if !isDerivedClientExpr(pass, call.Args[idx], allowedWrappers, derivedVars, factCache) {
			return false
		}
	}

	return true
}

func isApprovedSourceCall(pass *analysis.Pass, call *ast.CallExpr, allowedWrappers map[string]struct{}) bool {
	fnObj := calledFunction(pass, call)
	if fnObj == nil || fnObj.Pkg() == nil {
		return false
	}

	qName := qualifiedFuncName(fnObj)
	if _, ok := allowedWrappers[qName]; ok {
		return true
	}

	if fnObj.Pkg().Path() != clientsPkg {
		return false
	}

	switch fnObj.Name() {
	case "NewAPIClientFromSDKResource", "MaybeNewAPIClientFromFrameworkResource":
		return true
	default:
		return false
	}
}

func qualifiedFuncName(fn *types.Func) string {
	if fn.Pkg() == nil {
		return fn.Name()
	}
	return fn.Pkg().Path() + "." + fn.Name()
}

func isAPIClientPointer(t types.Type) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Name() == "APIClient" && obj.Pkg() != nil && obj.Pkg().Path() == clientsPkg
}

func toSet(values []string) map[string]struct{} {
	s := make(map[string]struct{}, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			continue
		}
		s[v] = struct{}{}
	}
	return s
}

func isInElasticsearchDir(filename string) bool {
	cleaned := filepath.ToSlash(filename)
	return strings.Contains(cleaned, "/internal/elasticsearch/")
}

func copyDerivedMap(in map[*types.Var]bool) map[*types.Var]bool {
	out := make(map[*types.Var]bool, len(in))
	maps.Copy(out, in)
	return out
}


func mergeDerivedState(dst, a, b map[*types.Var]bool) {
	keys := make(map[*types.Var]struct{}, len(a)+len(b))
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}

	for k := range dst {
		delete(dst, k)
	}

	for k := range keys {
		dst[k] = a[k] && b[k]
	}
}
