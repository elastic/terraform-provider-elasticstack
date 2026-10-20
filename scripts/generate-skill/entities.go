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
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// entityKind is the Terraform binding kind.
type entityKind int

const (
	kindResource entityKind = 1 << iota
	kindDataSource
)

func (k entityKind) has(other entityKind) bool { return k&other != 0 }

// entity holds the minimal metadata needed to build the index.
// The full content is the docs file copied verbatim into the skill.
type entity struct {
	Name            string // e.g. "elasticstack_elasticsearch_index"
	ShortName       string // e.g. "elasticsearch_index"
	Kinds           entityKind
	ResourceSummary string
	DataSrcSummary  string
}

// loadEntities discovers all entities from docs/resources and docs/data-sources.
// Each .md file becomes an entity; both dirs are walked so resources that also
// have a data source (and vice-versa) are merged into a single entry.
func loadEntities(docsDir string) ([]*entity, error) {
	byName := map[string]*entity{}

	if err := collectFromDir(byName, filepath.Join(docsDir, "resources"), kindResource); err != nil {
		return nil, err
	}
	if err := collectFromDir(byName, filepath.Join(docsDir, "data-sources"), kindDataSource); err != nil {
		return nil, err
	}

	out := make([]*entity, 0, len(byName))
	for _, e := range byName {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// collectFromDir reads every .md file in dir, parses minimal metadata, and
// creates or updates the matching entity in byName.
func collectFromDir(byName map[string]*entity, dir string, kind entityKind) error {
	des, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read docs dir %q: %w", dir, err)
	}

	for _, de := range des {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, de.Name()))
		if err != nil {
			return fmt.Errorf("read %q: %w", de.Name(), err)
		}
		meta := parseFrontmatter(extractFrontmatter(string(raw)))
		shortName := strings.TrimSuffix(de.Name(), ".md")
		fullName := "elasticstack_" + shortName

		ent, ok := byName[fullName]
		if !ok {
			ent = &entity{Name: fullName, ShortName: shortName}
			byName[fullName] = ent
		}
		ent.Kinds |= kind
		switch kind {
		case kindResource:
			ent.ResourceSummary = meta.Description
		case kindDataSource:
			ent.DataSrcSummary = meta.Description
		}
	}
	return nil
}

// extractFrontmatter returns the content between the leading --- delimiters,
// or an empty string if none is present.
func extractFrontmatter(body string) string {
	if !strings.HasPrefix(body, "---\n") {
		return ""
	}
	if end := strings.Index(body[4:], "\n---\n"); end >= 0 {
		return body[4 : 4+end]
	}
	return ""
}

type docsMeta struct {
	Subcategory string
	Description string
}

// parseFrontmatter handles the tiny subset of YAML that tfplugindocs emits:
// single-line key: value pairs and multi-line description with the |- indicator.
func parseFrontmatter(fm string) docsMeta {
	var meta docsMeta
	lines := strings.Split(fm, "\n")
	i := 0
	for i < len(lines) {
		line := lines[i]
		if after, ok := strings.CutPrefix(line, "subcategory:"); ok {
			v := strings.TrimSpace(after)
			meta.Subcategory = strings.Trim(v, `"'`)
		}
		if after, ok := strings.CutPrefix(line, "description:"); ok {
			v := strings.TrimSpace(after)
			if v == "|-" || v == "|" {
				var parts []string
				i++
				for i < len(lines) && (strings.HasPrefix(lines[i], "  ") || lines[i] == "") {
					parts = append(parts, strings.TrimSpace(lines[i]))
					i++
				}
				meta.Description = strings.TrimSpace(strings.Join(parts, " "))
				continue
			}
			meta.Description = strings.Trim(v, `"'`)
		}
		i++
	}
	return meta
}
