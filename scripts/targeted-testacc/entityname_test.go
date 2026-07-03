package main

import (
	"reflect"
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

func TestExtractFromSource_AllFourPatterns(t *testing.T) {
	src := `package mixed

func init() {
	_ = entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
	_ = entitycore.NewElasticsearchResource[Model]("index_template", opts)
	_ = entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "slo", opts)
	_ = entitycore.NewKibanaDataSource[Model](entitycore.ComponentKibana, "spaces", opts)
}`

	got := extractFromSource(src, nil)
	want := []EntityRef{
		{Component: "kibana", Name: "space"},
		{Component: "elasticsearch", Name: "index_template"},
		{Component: "kibana", Name: "slo"},
		{Component: "kibana", Name: "spaces"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractFromSource = %v, want %v", got, want)
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
