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
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestExtractFromSource_NewResourceBase(t *testing.T) {
	src := `package kibana

func init() {
	_ = entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
}`

	got, unresolved := extractFromSource(src)
	if len(unresolved) != 0 {
		t.Errorf("unexpected unresolved sites: %v", unresolved)
	}
	want := []EntityRef{{Component: "kibana", Name: "space"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestExtractFromSource_NewResourceBase_ShorthandForm(t *testing.T) {
	src := `package kibana

func init() {
	_ = NewResourceBase(ComponentKibana, "data_view")
}`

	got, _ := extractFromSource(src)
	want := []EntityRef{{Component: "kibana", Name: "data_view"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestExtractFromSource_NewElasticsearchResource(t *testing.T) {
	src := `package elasticsearch

func init() {
	_ = entitycore.NewElasticsearchResource[Model]("index_template", opts)
}`

	got, _ := extractFromSource(src)
	want := []EntityRef{{Component: "elasticsearch", Name: "index_template"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestExtractFromSource_NewKibanaResource(t *testing.T) {
	src := `package kibana

func init() {
	_ = entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "slo", opts)
}`

	got, _ := extractFromSource(src)
	want := []EntityRef{{Component: "kibana", Name: "slo"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestExtractFromSource_NewKibanaDataSource(t *testing.T) {
	src := `package kibana

func init() {
	_ = entitycore.NewKibanaDataSource[Model](entitycore.ComponentKibana, "spaces", opts)
}`

	got, _ := extractFromSource(src)
	want := []EntityRef{{Component: "kibana", Name: "spaces"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestExtractFromSource_AllCoveredConstructors(t *testing.T) {
	src := `package mixed

func init() {
	_ = entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
	_ = entitycore.NewDataSourceBase(entitycore.ComponentKibana, "data_view")
	_ = entitycore.NewEphemeralBase(entitycore.ComponentFleet, "agent")
	_ = entitycore.NewActionBase(entitycore.ComponentAPM, "source_map")
	_ = entitycore.NewElasticsearchResource[Model]("index_template", opts)
	_ = entitycore.NewElasticsearchDataSource[Model](entitycore.ComponentElasticsearch, "role", schema, read)
	_ = entitycore.NewElasticsearchEphemeralResource[Model, State]("apikey", opts)
	_ = entitycore.NewElasticsearchAction[Model]("snapshot_create", opts)
	_ = entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "slo", opts)
	_ = entitycore.NewKibanaDataSource[Model](entitycore.ComponentKibana, "spaces", opts)
	_ = entitycore.NewKibanaEphemeralResource[Model, State]("synthetic", opts)
	_ = entitycore.NewKibanaAction[Model]("bulk_upload", opts)
	// Type-inferred call sites (no explicit type argument):
	_ = entitycore.NewElasticsearchResource("synonym_set", opts)
	_ = entitycore.NewElasticsearchDataSource(entitycore.ComponentElasticsearch, "connector", schema, read)
}`

	got, unresolved := extractFromSource(src)
	if len(unresolved) != 0 {
		t.Errorf("unexpected unresolved sites: %v", unresolved)
	}
	want := []EntityRef{
		// entities are reported in source order
		{Component: "kibana", Name: "space"},
		{Component: "kibana", Name: "data_view"},
		{Component: "fleet", Name: "agent"},
		{Component: "apm", Name: "source_map"},
		{Component: "elasticsearch", Name: "index_template"},
		{Component: "elasticsearch", Name: "role"},
		{Component: "elasticsearch", Name: "apikey"},
		{Component: "elasticsearch", Name: "snapshot_create"},
		{Component: "kibana", Name: "slo"},
		{Component: "kibana", Name: "spaces"},
		{Component: "kibana", Name: "synthetic"},
		{Component: "kibana", Name: "bulk_upload"},
		// type-inferred call sites (no explicit type argument)
		{Component: "elasticsearch", Name: "synonym_set"},
		{Component: "elasticsearch", Name: "connector"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

// TestEntitycoreConstructorsCovered guards against constructor drift: every
// exported New* constructor in internal/entitycore must be classified either
// as a covered entity constructor (extracted by phase 2) or as a non-entity
// constructor. A new envelope type added to entitycore without a matching
// extraction pattern fails this test.
func TestEntitycoreConstructorsCovered(t *testing.T) {
	root := repoRoot(t)
	entitycoreDir := filepath.Join(root, "internal/entitycore")

	covered := make(map[string]bool, len(coveredEntityConstructors)+len(nonEntityConstructors))
	for _, name := range coveredEntityConstructors {
		covered[name] = true
	}
	for _, name := range nonEntityConstructors {
		covered[name] = true
	}
	coveredRE := regexp.MustCompile(`^func (New\w+)[\(\[]`)

	// Walk recursively: a constructor added under a future entitycore
	// subpackage must trip this guard too, not evade it.
	var goFiles []string
	err := filepath.WalkDir(entitycoreDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		goFiles = append(goFiles, path)
		return nil
	})
	if err != nil {
		// A skip here would let the guard silently pass; fail instead.
		t.Fatalf("cannot walk %s: %v", entitycoreDir, err)
	}

	var unclassified []string
	for _, path := range goFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for line := range strings.SplitSeq(string(data), "\n") {
			m := coveredRE.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if !covered[m[1]] {
				unclassified = append(unclassified, m[1])
			}
		}
	}

	if len(unclassified) > 0 {
		t.Errorf("internal/entitycore exports New* constructors not classified by scripts/targeted-testacc: %v\n"+
			"Add the constructor to coveredEntityConstructors (with an extraction pattern in entityname.go) "+
			"or to nonEntityConstructors in the same file.", unclassified)
	}
}

// TestExtractFromSource_IgnoresCommentedCallSites asserts that entities
// are extracted via the go/ast AST, so constructor invocations written in
// comments (line comments and block comments alike) are ignored naturally.
func TestExtractFromSource_IgnoresCommentedCallSites(t *testing.T) {
	src := `package mixed

// entitycore.NewResourceBase(entitycore.ComponentKibana, "ignored")
func init() {
	_ = entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
	/*
		entitycore.NewElasticsearchResource[Model]("ignored", opts)
	*/
}`

	got, _ := extractFromSource(src)
	want := []EntityRef{{Component: "kibana", Name: "space"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
	}
}

func TestComponentName(t *testing.T) {
	cases := []struct {
		suffix string
		want   string
		ok     bool
	}{
		{"Elasticsearch", "elasticsearch", true},
		{"Kibana", "kibana", true},
		{"Fleet", "fleet", true},
		{"APM", "apm", true},
		{"Unknown", "", false},
		{"", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.suffix, func(t *testing.T) {
			got, ok := componentName(tc.suffix)
			if got != tc.want || ok != tc.ok {
				t.Errorf("componentName(%q) = (%q, %v), want (%q, %v)", tc.suffix, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestEntityRef_FullName(t *testing.T) {
	cases := []struct {
		ref  EntityRef
		want string
	}{
		{EntityRef{Component: "kibana", Name: "space"}, "elasticstack_kibana_space"},
		{EntityRef{Component: "elasticsearch", Name: "index_template"}, "elasticstack_elasticsearch_index_template"},
		{EntityRef{Component: "fleet", Name: "agent_policy"}, "elasticstack_fleet_agent_policy"},
		{EntityRef{Component: "apm", Name: "source_map"}, "elasticstack_apm_source_map"},
	}

	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.ref.FullName(); got != tc.want {
				t.Errorf("FullName() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestExtractEntities_FromRepoCallSites asserts that the AST extraction
// still matches the real constructor call sites in the repository, rather
// than only synthetic fixtures. The sampled directories below must each
// yield the named entities, so an extraction that stops matching the tree
// fails loudly here.
//
// The repo-wide floor below covers the packages the samples cannot: a
// regression in any unsampled package would drop the extracted-entity
// totals under the recorded floor. The floor counts unique entities across
// every enumerated acceptance-test package (measured on the rebase that
// moved extraction from regexes to the go/ast AST: 136 packages yield 100
// entity references, 96 unique); the floor deliberately stays below today's
// count so that legitimate single-package additions do not need a test
// update, while a systematic extraction regression cannot pass unnoticed.
func TestExtractEntities_FromRepoCallSites(t *testing.T) {
	cases := []struct {
		dir  string
		want []string
	}{
		{"internal/elasticsearch/synonyms", []string{
			"elasticstack_elasticsearch_synonym_set",
		}},
		{"internal/elasticsearch/queryrulesets", []string{
			"elasticstack_elasticsearch_query_ruleset",
		}},
		{"internal/elasticsearch/connector/resource", []string{
			"elasticstack_elasticsearch_connector",
		}},
		{"internal/elasticsearch/security/role", []string{
			"elasticstack_elasticsearch_security_role",
		}},
		{"internal/kibana/spaces", []string{
			"elasticstack_kibana_space",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			got, err := ExtractEntities(filepath.Join("..", "..", tc.dir))
			if err != nil {
				t.Fatalf("ExtractEntities(%s): %v", tc.dir, err)
			}
			names := make([]string, 0, len(got))
			for _, e := range got {
				names = append(names, e.FullName())
			}
			if len(names) == 0 {
				t.Fatalf("ExtractEntities(%s) extracted no entities; extraction no longer matches real call sites", tc.dir)
			}
			for _, want := range tc.want {
				if !slices.Contains(names, want) {
					t.Errorf("ExtractEntities(%s) missing %s; got %v", tc.dir, want, names)
				}
			}
		})
	}
}

// TestExtractEntities_AcrossAllAccPackages iterates every enumerated
// acceptance-test package and asserts a whole-repo floor on extracted
// entities. Per-package "must declare an entity" is not sound: many
// enumerated packages legitimately declare none (the internal/kibana/dashboard/panel/*
// fixture packages, internal/acctest, internal/clients, provider and other
// helper packages declare their entities in the packages they test). A
// unique-entity floor over the whole set still catches a systematic
// extraction regression across all 136 packages.
func TestExtractEntities_AcrossAllAccPackages(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(root)
	modulePath, err := currentModulePath()
	if err != nil {
		t.Fatalf("cannot resolve module path: %v", err)
	}

	pkgs, err := FindAccTestPackages(accTestEnumerationRoots, modulePath)
	if err != nil {
		t.Fatalf("FindAccTestPackages: %v", err)
	}
	if len(pkgs) < 100 {
		t.Errorf("enumerated only %d acceptance test packages (expected ≥100; enumeration regression?)", len(pkgs))
	}

	unique := make(map[string]EntityRef)
	refs := 0
	for _, pkg := range pkgs {
		dir := strings.TrimPrefix(pkg, modulePath+"/")
		ents, err := ExtractEntities(dir)
		if err != nil {
			t.Errorf("ExtractEntities(%s): %v", dir, err)
			continue
		}
		refs += len(ents)
		for _, e := range ents {
			unique[e.FullName()] = e
		}
	}

	// Floor derived from the measured 2025-11-16 count: 136 packages, 100
	// references, 96 unique entities. Keeps headroom for entity additions
	// without test updates; set from 96 to catch any broad regression.
	const floorUnique = 96
	if len(unique) < floorUnique {
		sample := make([]string, 0, 10)
		for name := range unique {
			sample = append(sample, name)
		}
		sort.Strings(sample)
		if len(sample) > 10 {
			sample = sample[:10]
		}
		t.Errorf("extraction over %d acceptance test packages yielded only %d unique entities (floor %d; %d refs). Sample: %v",
			len(pkgs), len(unique), floorUnique, refs, sample)
	}
}

// TestExtractUnresolvedSites_RepoWide fails when a constructor call site
// under the tool's enumeration roots cannot have its entity name resolved
// statically. The only legitimately unresolved sites today are entitycore's
// own envelope constructors forwarding their (component, name) parameters;
// any new one — a call site building its entity name from a variable or by
// concatenation — makes that entity invisible to the selector and must be
// either rewritten at the call site to a plain string literal or handled
// explicitly here.
func TestExtractUnresolvedSites_RepoWide(t *testing.T) {
	root := repoRoot(t)
	t.Chdir(root)

	// allowed documents the known unresolved sites, each with the reason it
	// is acceptable.
	allowed := map[string]string{
		"internal/entitycore/action_envelope.go":                  "envelope constructor forwards its (component, name) parameters to NewActionBase",
		"internal/entitycore/data_source_envelope.go":             "envelope constructor forwards its (component, name) parameters to NewDataSourceBase",
		"internal/entitycore/elasticsearch_ephemeral_envelope.go": "envelope constructor forwards its name parameter to NewEphemeralBase",
		"internal/entitycore/kibana_ephemeral_envelope.go":        "envelope constructor forwards its name parameter to NewEphemeralBase",
		"internal/entitycore/kibana_resource_envelope.go":         "envelope constructor forwards its (component, name) parameters to NewResourceBase",
		"internal/entitycore/resource_envelope.go":                "envelope constructor forwards its name parameter to NewResourceBase",
	}

	var unexpected []string
	for _, top := range accTestEnumerationRoots {
		err := filepath.WalkDir(top, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if d.Name() == "testdata" {
					return fs.SkipDir
				}
				return nil
			}
			name := d.Name()
			if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			_, sites := extractFromFile(path, data)
			for _, site := range sites {
				rel := filepath.ToSlash(path)
				if _, ok := allowed[rel]; ok {
					continue
				}
				unexpected = append(unexpected, fmt.Sprintf("%s:%d %s", rel, site.Line, site.Constructor))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", top, err)
		}
	}

	if len(unexpected) > 0 {
		t.Errorf("covered constructor call sites with non-literal entity names (invisible to the selector):\n%s\n"+
			"Rewrite the call site to pass a plain string literal, or document it in this test's allowlist.",
			strings.Join(unexpected, "\n"))
	}
}

func TestExtractEntities_DeduplicatesAndSorts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "resource.go", `
package slo

func init() {
	_ = NewResourceBase(ComponentKibana, "slo")
}
`)
	writeFile(t, root, "resource2.go", `
package slo

func init() {
	_ = NewResourceBase(ComponentKibana, "slo")
	_ = NewResourceBase(ComponentKibana, "space")
}
`)

	got, err := ExtractEntities(root)
	if err != nil {
		t.Fatalf("ExtractEntities: %v", err)
	}
	want := []EntityRef{
		{Component: "kibana", Name: "slo"},
		{Component: "kibana", Name: "space"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractEntities = %v, want %v", got, want)
	}
}

func TestExtractEntities_SkipsTestFiles(t *testing.T) {
	root := t.TempDir()
	// A resource declared in the package source.
	writeFile(t, root, "resource.go", `
package slo

func init() {
	_ = NewResourceBase(ComponentKibana, "slo")
}
`)
	// A phantom entity declared only in a test file; it must be ignored
	// because _test.go files are excluded from entity extraction.
	writeFile(t, root, "resource_test.go", `
package slo_test

func init() {
	_ = NewResourceBase(ComponentKibana, "space")
}
`)

	got, err := ExtractEntities(root)
	if err != nil {
		t.Fatalf("ExtractEntities: %v", err)
	}
	want := []EntityRef{{Component: "kibana", Name: "slo"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExtractEntities = %v, want %v", got, want)
	}
}
