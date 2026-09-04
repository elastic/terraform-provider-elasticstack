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
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestExtractFromSource_NewResourceBase(t *testing.T) {
	src := `package kibana

func init() {
	_ = entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
}`

	got := extractFromSource(src, nil)
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

	got := extractFromSource(src, nil)
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

	got := extractFromSource(src, nil)
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

	got := extractFromSource(src, nil)
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

	got := extractFromSource(src, nil)
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
	_ = entitycore.NewElasticsearchDataSource[Model]("role", schema, read)
	_ = entitycore.NewElasticsearchEphemeralResource[Model, State]("apikey", opts)
	_ = entitycore.NewElasticsearchAction[Model]("snapshot_create", opts)
	_ = entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "slo", opts)
	_ = entitycore.NewKibanaDataSource[Model](entitycore.ComponentKibana, "spaces", opts)
	_ = entitycore.NewKibanaEphemeralResource[Model, State]("synthetic", opts)
	_ = entitycore.NewKibanaAction[Model]("bulk_upload", opts)
}`

	got := extractFromSource(src, nil)
	want := []EntityRef{
		// baseEntityRE
		{Component: "kibana", Name: "space"},
		{Component: "kibana", Name: "data_view"},
		{Component: "fleet", Name: "agent"},
		{Component: "apm", Name: "source_map"},
		// kibanaComponentRE
		{Component: "kibana", Name: "slo"},
		{Component: "kibana", Name: "spaces"},
		// elasticsearchNameRE
		{Component: "elasticsearch", Name: "index_template"},
		{Component: "elasticsearch", Name: "role"},
		{Component: "elasticsearch", Name: "apikey"},
		{Component: "elasticsearch", Name: "snapshot_create"},
		// kibanaNameRE
		{Component: "kibana", Name: "synthetic"},
		{Component: "kibana", Name: "bulk_upload"},
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
	const entitycoreDir = "../../internal/entitycore"

	entries, err := os.ReadDir(entitycoreDir)
	if err != nil {
		t.Skipf("cannot read %s: %v (not running inside the repository?)", entitycoreDir, err)
	}

	covered := make(map[string]bool, len(coveredEntityConstructors)+len(nonEntityConstructors))
	for _, name := range coveredEntityConstructors {
		covered[name] = true
	}
	for _, name := range nonEntityConstructors {
		covered[name] = true
	}
	coveredRE := regexp.MustCompile(`^func (New\w+)\(`)

	var unclassified []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(entitycoreDir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
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

func TestExtractFromSource_HonoursCommentIntervals(t *testing.T) {
	src := `package mixed

// entitycore.NewResourceBase(entitycore.ComponentKibana, "ignored")
func init() {
	/*
		entitycore.NewElasticsearchResource[Model]("ignored", opts)
	*/
}`

	// Simulate intervals covering the entire source; extractFromSource should
	// ignore every match.
	intervals := [][2]int{{0, 200}}

	got := extractFromSource(src, intervals)
	if len(got) != 0 {
		t.Errorf("expected no entities, got %v", got)
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
