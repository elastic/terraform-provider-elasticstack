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

package acctest

import (
	"io/fs"
	"sort"
	"testing"
	"testing/fstest"
)

func TestExamplesHarness_collectTfExamples(t *testing.T) {
	t.Parallel()

	cfgRes := []byte(`provider "elasticstack" {}

resource "elasticstack_elasticsearch_index" "i" {
  name                = "n"
  deletion_protection = false
}
`)
	cfgDSOnly := []byte(`provider "elasticstack" {}

data "elasticstack_elasticsearch_info" "x" {}
`)
	cfgDSMixed := []byte(`provider "elasticstack" {}

resource "elasticstack_elasticsearch_index" "y" {}

data "elasticstack_elasticsearch_enrich_policy" "p" {
  name = "n"
}
`)

	mfs := fstest.MapFS{
		"resources/r1/resource.tf":              &fstest.MapFile{Data: cfgRes, Mode: 0o644},
		"data-sources/ds1/data.tf":              &fstest.MapFile{Data: cfgDSOnly, Mode: 0o644},
		"data-sources/ds2/data.tf":              &fstest.MapFile{Data: cfgDSMixed, Mode: 0o644},
		"resources/ignored-non-tf-manifest.txt": &fstest.MapFile{Data: []byte("x"), Mode: 0o644},
	}

	got, err := collectTfExamples(mfs)
	if err != nil {
		t.Fatalf("collectTfExamples: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(cases) = %d, want 3 (non-.tf skipped)", len(got))
	}
	sort.Slice(got, func(i, j int) bool {
		return got[i].pathUnderExamples < got[j].pathUnderExamples
	})

	want := []struct {
		pathUnder      string
		repoExamples   string
		expectNonEmpty bool
	}{
		{"data-sources/ds1/data.tf", "examples/data-sources/ds1/data.tf", false},
		{"data-sources/ds2/data.tf", "examples/data-sources/ds2/data.tf", true},
		{"resources/r1/resource.tf", "examples/resources/r1/resource.tf", true},
	}

	for i := range got {
		if got[i].pathUnderExamples != want[i].pathUnder {
			t.Fatalf("[%d].pathUnderExamples = %q, want %q", i, got[i].pathUnderExamples, want[i].pathUnder)
		}
		if got[i].repoExamplesPath != want[i].repoExamples {
			t.Fatalf("[%d].repoExamplesPath = %q, want %q", i, got[i].repoExamplesPath, want[i].repoExamples)
		}
		if got[i].embedRelativePath != want[i].pathUnder {
			t.Fatalf("[%d].embedRelativePath = %q, want %q", i, got[i].embedRelativePath, want[i].pathUnder)
		}
		body, errFS := fs.ReadFile(mfs, got[i].embedRelativePath)
		if errFS != nil {
			t.Fatalf("[%d] ReadFile: %v", i, errFS)
		}
		if en := expectNonEmptyPlanForExample(got[i].repoExamplesPath, body); en != want[i].expectNonEmpty {
			t.Fatalf("[%d] expectNonEmptyPlanForExample = %v, want %v", i, en, want[i].expectNonEmpty)
		}
	}
}

func TestExamplesHarness_tfRootDeclaresResourceOrOutput(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		cfg  string
		want bool
	}{
		{
			name: "top_level_resource",
			cfg: `provider "elasticstack" {}

resource "elasticstack_elasticsearch_index" "x" {
  name = "i"
}`,
			want: true,
		},
		{
			name: "top_level_output",
			cfg: `provider "elasticstack" {}

output "x" {
  value = "y"
}`,
			want: true,
		},
		{
			name: "data_only_read",
			cfg: `provider "elasticstack" {}

data "elasticstack_elasticsearch_info" "cluster" {}`,
			want: false,
		},
		{
			name: "comment_looks_like_resource",
			cfg: `# resource "x" "y" {}

provider "elasticstack" {}

data "elasticstack_elasticsearch_info" "cluster" {}`,
			want: false,
		},
		{
			name: "heredoc_contains_resource_syntax",
			cfg: `provider "elasticstack" {}

data "elasticstack_elasticsearch_info" "cluster" {}

locals {
  tmpl = <<-EOT
resource "bogus" "n" {}
EOT
}`,
			want: false,
		},
		{
			name: "parse_error_returns_false_conservative_expectation",
			cfg:  `this is not valid hcl {`,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tfRootDeclaresResourceOrOutput([]byte(tc.cfg)); got != tc.want {
				t.Fatalf("tfRootDeclaresResourceOrOutput(...) = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestExamplesHarness_expectNonEmptyPlanForExample(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		repoExamples   string
		cfg            string
		wantNonEmptyPN bool
	}{
		{
			name:           "resources_tree_always_true",
			repoExamples:   "examples/resources/elasticstack_elasticsearch_index/resource.tf",
			cfg:            `provider "elasticstack" {}`,
			wantNonEmptyPN: true,
		},
		{
			name:           "data_sources_read_only_false",
			repoExamples:   "examples/data-sources/elasticstack_elasticsearch_info/data-source.tf",
			cfg:            `provider "elasticstack" {}\ndata "elasticstack_elasticsearch_info" "x" {}`,
			wantNonEmptyPN: false,
		},
		{
			name:         "data_sources_with_resource_true",
			repoExamples: "examples/data-sources/elasticstack_elasticsearch_enrich_policy/data-source.tf",
			cfg: `provider "elasticstack" {}

resource "elasticstack_elasticsearch_index" "y" {}

data "elasticstack_elasticsearch_enrich_policy" "p" {
  name = "n"
}`,
			wantNonEmptyPN: true,
		},
		{
			name:         "data_sources_with_output_true",
			repoExamples: "examples/data-sources/elasticstack_elasticsearch_snapshot_repository/data-source.tf",
			cfg: `provider "elasticstack" {}

output "x" {
  value = "y"
}

data "elasticstack_elasticsearch_snapshot_repository" "r" {}`,
			wantNonEmptyPN: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := expectNonEmptyPlanForExample(tc.repoExamples, []byte(tc.cfg)); got != tc.wantNonEmptyPN {
				t.Fatalf("expectNonEmptyPlanForExample(...) = %v, want %v", got, tc.wantNonEmptyPN)
			}
		})
	}
}

func TestExamplesHarness_shouldSkipExamplePath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want bool
	}{
		{"examples/cloud/foo.tf", true},
		{"examples/provider/snippet.tf", true},
		{"examples/resources/elasticstack_elasticsearch_index/resource.tf", false},
		{"examples/data-sources/elasticstack_elasticsearch_info/data-source.tf", false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			if got := shouldSkipExamplePath(tc.path); got != tc.want {
				t.Fatalf("shouldSkipExamplePath(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}
