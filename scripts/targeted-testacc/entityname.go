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

package main

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

// EntityRef identifies a Terraform resource/data source declared in Go source.
type EntityRef struct {
	Component string
	Name      string
}

// FullName returns the Terraform type name, e.g. elasticstack_kibana_space.
func (e EntityRef) FullName() string {
	return fmt.Sprintf("elasticstack_%s_%s", e.Component, e.Name)
}

// coveredEntityConstructors lists every entitycore constructor that declares
// a Terraform entity name and is covered by the AST extraction below. The
// guard test TestEntitycoreConstructorsCovered fails when entitycore exports
// a New* constructor that is not classified here, so future envelope types
// cannot silently produce phase-2 false negatives.
var coveredEntityConstructors = []string{
	"NewActionBase",
	"NewDataSourceBase",
	"NewElasticsearchAction",
	"NewElasticsearchDataSource",
	"NewElasticsearchEphemeralResource",
	"NewElasticsearchResource",
	"NewEphemeralBase",
	"NewKibanaAction",
	"NewKibanaDataSource",
	"NewKibanaEphemeralResource",
	"NewKibanaResource",
	"NewResourceBase",
}

// nonEntityConstructors lists exported entitycore constructors that do not
// declare a Terraform entity name (importers, version requirements), so they
// are not part of the extraction table.
var nonEntityConstructors = []string{
	"NewAttributeVersionCheckRequirement",
	"NewAttributeVersionRequirement",
	"NewCompositeIDImporter",
	"NewKibanaSpaceImporter",
	"NewSpaceImporter",
}

// constructorArgKind classifies how a covered constructor declares its entity
// name and component in its argument list.
type constructorArgKind int

const (
	// argKindComponentFirst constructors take (component, name, ...) with a
	// mandatory Component<X> first argument:
	//   entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
	//   entitycore.NewEphemeralBase(entitycore.ComponentFleet, "agent")
	//   entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "slo", opts)
	// The entitycore. and Component prefixes are optional at the call site:
	// both a dot-imported shorthand and an aliased import must extract.
	argKindComponentFirst constructorArgKind = iota

	// argKindElasticsearchNameFirst constructors take ("name", ...) with an
	// optional leading Component<X> argument that overrides the default
	// Elasticsearch component:
	//   entitycore.NewElasticsearchResource[Model]("index_template", opts)
	//   entitycore.NewElasticsearchDataSource[Model](entitycore.ComponentElasticsearch, "role", schema, read)
	argKindElasticsearchNameFirst

	// argKindKibanaNameFirst constructors take ("name", ...) with an
	// optional leading Component<X> argument that overrides the default
	// Kibana component:
	//   entitycore.NewKibanaEphemeralResource[Model, State]("synthetic", opts)
	//   entitycore.NewKibanaAction[Model]("bulk_upload", opts)
	argKindKibanaNameFirst
)

// coveredConstructorArgs maps every constructor in coveredEntityConstructors
// to its argument shape. TestCoveredConstructorClassification keeps this
// table in sync with coveredEntityConstructors.
var coveredConstructorArgs = map[string]constructorArgKind{
	"NewActionBase":                     argKindComponentFirst,
	"NewDataSourceBase":                 argKindComponentFirst,
	"NewElasticsearchAction":            argKindElasticsearchNameFirst,
	"NewElasticsearchDataSource":        argKindElasticsearchNameFirst,
	"NewElasticsearchEphemeralResource": argKindElasticsearchNameFirst,
	"NewElasticsearchResource":          argKindElasticsearchNameFirst,
	"NewEphemeralBase":                  argKindComponentFirst,
	"NewKibanaAction":                   argKindKibanaNameFirst,
	"NewKibanaDataSource":               argKindComponentFirst,
	"NewKibanaEphemeralResource":        argKindKibanaNameFirst,
	"NewKibanaResource":                 argKindComponentFirst,
	"NewResourceBase":                   argKindComponentFirst,
}

const (
	componentElasticsearch = "elasticsearch"
	componentKibana        = "kibana"
	componentFleet         = "fleet"
	componentAPM           = "apm"
)

// componentName maps the Go identifier suffix (e.g. "Kibana", "APM") to the
// string value used in Terraform type names.
func componentName(suffix string) (string, bool) {
	switch suffix {
	case "Elasticsearch":
		return componentElasticsearch, true
	case "Kibana":
		return componentKibana, true
	case "Fleet":
		return componentFleet, true
	case "APM":
		return componentAPM, true
	}
	return "", false
}

// unresolvedEntitySite records a covered entitycore constructor call site
// whose entity name (or component argument) cannot be statically resolved —
// for example a call forwarding a parameter or building the name by
// concatenation instead of passing a plain string literal. Such a site is
// distinct from "no entity here": the entity may well exist, but this tool
// cannot know its name, so silently dropping it would hide a selection gap.
// Extraction records these sites instead of guessing; the guard test
// TestExtractUnresolvedSites_RepoWide fails when one appears outside the
// documented internal/entitycore delegation sites.
type unresolvedEntitySite struct {
	Constructor string
	Filename    string
	Line        int
}

// ExtractEntities scans all non-test .go files (files ending in _test.go are
// excluded to avoid phantom entities from string literals in test source) in
// dir and returns the unique set of Terraform entities declared in that
// package directory. Call sites whose entity name cannot be resolved
// statically are skipped (mirroring the historical behaviour) but are
// recorded as unresolvedEntitySite by extractEntities and guarded repo-wide
// by TestExtractUnresolvedSites_RepoWide.
func ExtractEntities(dir string) ([]EntityRef, error) {
	entities := make(map[string]EntityRef)
	var unresolved []unresolvedEntitySite

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}

		ents, sites := extractFromFile(path, data)
		for _, ent := range ents {
			entities[ent.FullName()] = ent
		}
		unresolved = append(unresolved, sites...)
	}

	result := make([]EntityRef, 0, len(entities))
	for _, ent := range entities {
		result = append(result, ent)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FullName() < result[j].FullName()
	})
	_ = unresolved // consumed by tests via extractEntities
	return result, nil
}

// extractFromFile parses a single Go source file and returns all entity
// references found in it, together with any unresolved constructor call
// sites. Comments are excluded by construction: the AST never contains call
// expressions written in comments. Files that cannot be parsed cannot
// contribute valid entity declarations and are skipped.
func extractFromFile(filename string, data []byte) ([]EntityRef, []unresolvedEntitySite) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, data, parser.ParseComments|parser.AllErrors)
	if err != nil {
		return nil, nil
	}
	return extractFromAST(fset, f)
}

// extractFromSource parses a single Go source string and returns all entity
// references found, together with any unresolved constructor call sites. The
// function is exposed independently so unit tests can pass synthetic source
// snippets.
func extractFromSource(src string) ([]EntityRef, []unresolvedEntitySite) {
	return extractFromFile("<source>", []byte(src))
}

// extractFromAST walks a parsed file and collects entity declarations from
// calls to covered entitycore constructors.
func extractFromAST(fset *token.FileSet, f *ast.File) ([]EntityRef, []unresolvedEntitySite) {
	var entities []EntityRef
	var unresolved []unresolvedEntitySite

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ctor := coveredConstructorName(call.Fun)
		if ctor == "" {
			return true
		}

		ref, resolved := entityRefFromCall(ctor, call.Args)
		if !resolved {
			pos := fset.Position(call.Pos())
			unresolved = append(unresolved, unresolvedEntitySite{
				Constructor: ctor,
				Filename:    pos.Filename,
				Line:        pos.Line,
			})
			return true
		}
		if ref.Name != "" {
			entities = append(entities, ref)
		}
		return true
	})

	return entities, unresolved
}

// coveredConstructorName reports the covered constructor name for a call
// expression's function operand, unwrapping explicit type-argument lists
// (NewKibanaResource[Model](...)). The entitycore. qualifier is optional —
// and any receiver (an aliased import, e.g. ec.NewResourceBase) is accepted,
// mirroring the historical textual scan.
func coveredConstructorName(fun ast.Expr) string {
	var name string
	switch fn := fun.(type) {
	case *ast.Ident:
		name = fn.Name
	case *ast.SelectorExpr:
		name = fn.Sel.Name
	case *ast.IndexExpr:
		return coveredConstructorName(fn.X)
	case *ast.IndexListExpr:
		return coveredConstructorName(fn.X)
	default:
		return ""
	}
	if _, ok := coveredConstructorArgs[name]; ok {
		return name
	}
	return ""
}

// entityRefFromCall classifies the arguments of a covered constructor call
// and returns the entity it declares. resolved is true when the outcome is
// definitively known either way — the call declares the returned entity, or
// it declares none (missing name argument, unrecognised component). resolved
// is false when the call plausibly declares an entity but the name or the
// component is not statically resolvable (parameter forwarding,
// concatenation, a name-first spelling of a component-first constructor):
// the caller records such sites instead of silently dropping them.
func entityRefFromCall(ctor string, args []ast.Expr) (ref EntityRef, resolved bool) {
	kind := coveredConstructorArgs[ctor]
	if len(args) == 0 {
		// No arguments at all: nothing that could declare an entity name.
		return EntityRef{}, true
	}

	componentSuffix := ""
	hasComponent := false
	nameIdx := 0
	if suffix, ok := componentExprSuffix(args[0]); ok {
		hasComponent = true
		componentSuffix = suffix
		nameIdx = 1
	}

	if !hasComponent {
		switch kind {
		case argKindComponentFirst:
			// The component argument is mandatory. Without it the position
			// of the name cannot be determined statically: either the call
			// forwards variables (internal/entitycore's own delegation) or it
			// is spelled name-first. Record, never guess.
			if len(args) >= 2 {
				return EntityRef{}, false
			}
			return EntityRef{}, true
		case argKindElasticsearchNameFirst:
			componentSuffix = "Elasticsearch"
		case argKindKibanaNameFirst:
			componentSuffix = "Kibana"
		}
	}

	if nameIdx >= len(args) {
		// A component argument without a following name declares nothing.
		return EntityRef{}, true
	}

	name, ok := stringLiteral(args[nameIdx])
	if !ok {
		return EntityRef{}, false
	}

	component, ok := componentName(componentSuffix)
	if !ok {
		// A component outside the recognised set (ComponentFoo) declares
		// nothing, mirroring the historical scan.
		return EntityRef{}, true
	}
	return EntityRef{Component: component, Name: name}, true
}

// componentExprSuffix returns the component suffix ("Kibana", "APM", ...)
// of a Component<X> argument, written either as the shorthand identifier
// (ComponentKibana), qualified (entitycore.ComponentKibana), or through an
// aliased import (ec.ComponentKibana).
func componentExprSuffix(e ast.Expr) (string, bool) {
	var name string
	switch x := e.(type) {
	case *ast.Ident:
		name = x.Name
	case *ast.SelectorExpr:
		name = x.Sel.Name
	default:
		return "", false
	}
	if suffix, ok := strings.CutPrefix(name, "Component"); ok && suffix != "" {
		return suffix, true
	}
	return "", false
}

// stringLiteral returns the value of a plain Go string literal argument,
// reporting false for anything else (identifiers, concatenations,
// conversions), so that non-literal names are never silently guessed.
func stringLiteral(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil || value == "" {
		return "", false
	}
	return value, true
}
