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
	"testing"
)

func TestParseDocsFrontmatter(t *testing.T) {
	body := `---
subcategory: "Security"
page_title: "something"
description: |-
  Adds and updates roles in the native realm. See the role API documentation https://example for more details.
---

# Something
`
	meta := parseFrontmatter(extractFrontmatter(body))
	if meta.Subcategory != "Security" {
		t.Errorf("Subcategory = %q", meta.Subcategory)
	}
	if meta.Description == "" {
		t.Errorf("Description empty")
	}
}

func TestLoadEntities(t *testing.T) {
	// Build a minimal fake docs tree.
	dir := t.TempDir()
	resDir := filepath.Join(dir, "resources")
	dsDir := filepath.Join(dir, "data-sources")
	if err := os.MkdirAll(resDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	writeDoc := func(path, sub, desc string) {
		t.Helper()
		content := "---\nsubcategory: \"" + sub + "\"\ndescription: |-\n  " + desc + "\n---\n# foo\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeDoc(filepath.Join(resDir, "elasticsearch_index.md"), "Elasticsearch", "Manages an index.")
	writeDoc(filepath.Join(resDir, "kibana_space.md"), "Kibana", "Manages a space.")
	writeDoc(filepath.Join(dsDir, "elasticsearch_index.md"), "Elasticsearch", "Reads an index.")

	entities, err := loadEntities(dir)
	if err != nil {
		t.Fatal(err)
	}

	byName := map[string]*entity{}
	for _, e := range entities {
		byName[e.Name] = e
	}

	// elasticsearch_index should be both resource and data source.
	idx, ok := byName["elasticstack_elasticsearch_index"]
	if !ok {
		t.Fatal("elasticstack_elasticsearch_index not found")
	}
	if !idx.Kinds.has(kindResource) || !idx.Kinds.has(kindDataSource) {
		t.Errorf("elasticsearch_index kinds = %d, want both", idx.Kinds)
	}
	// kibana_space should be resource only.
	ks, ok := byName["elasticstack_kibana_space"]
	if !ok {
		t.Fatal("elasticstack_kibana_space not found")
	}
	if !ks.Kinds.has(kindResource) || ks.Kinds.has(kindDataSource) {
		t.Errorf("kibana_space kinds = %d, want resource only", ks.Kinds)
	}
}

func TestOneLineStripsSeeURL(t *testing.T) {
	in := "Creates Elasticsearch indices. See: https://example/docs"
	want := "Creates Elasticsearch indices."
	if got := oneLine(in); got != want {
		t.Errorf("oneLine = %q, want %q", got, want)
	}
}
