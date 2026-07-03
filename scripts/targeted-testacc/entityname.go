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
	"regexp"
	"sort"
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

var (
	// entitycore.NewResourceBase(entitycore.ComponentKibana, "space")
	// or NewResourceBase(ComponentKibana, "space")
	// or NewResourceBase(ComponentAPM, "source_map")
	newResourceBaseRE = regexp.MustCompile(`(?:entitycore\.)?NewResourceBase\s*\(\s*(?:entitycore\.)?Component(\w+)\s*,\s*"([^"]+)"\s*\)`)

	// entitycore.NewElasticsearchResource[Model]("name", opts)
	newElasticsearchResourceRE = regexp.MustCompile(`(?:entitycore\.)?NewElasticsearchResource\[[^\]]*\]\s*\(\s*"([^"]+)"`)

	// entitycore.NewKibanaResource[Model](entitycore.ComponentKibana, "name", opts)
	newKibanaResourceRE = regexp.MustCompile(`(?:entitycore\.)?NewKibanaResource\[[^\]]*\]\s*\(\s*(?:entitycore\.)?Component(\w+)\s*,\s*"([^"]+)"`)

	// entitycore.NewKibanaDataSource[Model](entitycore.ComponentKibana, "name", opts)
	newKibanaDataSourceRE = regexp.MustCompile(`(?:entitycore\.)?NewKibanaDataSource\[[^\]]*\]\s*\(\s*(?:entitycore\.)?Component(\w+)\s*,\s*"([^"]+)"`)
)

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

// ExtractEntities scans all .go files in dir and returns the unique set of
// Terraform entities declared in that package directory.
func ExtractEntities(dir string) ([]EntityRef, error) {
	entities := make(map[string]EntityRef)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", path, err)
		}

		ents := extractFromFile(path, data)
		for _, ent := range ents {
			entities[ent.FullName()] = ent
		}
	}

	result := make([]EntityRef, 0, len(entities))
	for _, ent := range entities {
		result = append(result, ent)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].FullName() < result[j].FullName()
	})
	return result, nil
}

// extractFromFile parses a single Go source file and returns all entity
// references found, ignoring matches inside comments.
func extractFromFile(filename string, data []byte) []EntityRef {
	src := string(data)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments|parser.AllErrors)
	if err != nil {
		// Skip files that cannot be parsed; they cannot contribute valid entity declarations.
		return nil
	}

	intervals := commentIntervals(fset, f.Comments)
	return extractFromSource(src, intervals)
}

// commentIntervals converts comment groups into [start, end) byte intervals.
// It uses the FileSet so that offsets are 0-based byte positions matching
// the regex locations used by extractFromSource.
func commentIntervals(fset *token.FileSet, groups []*ast.CommentGroup) [][2]int {
	intervals := make([][2]int, 0, len(groups))
	for _, g := range groups {
		if len(g.List) == 0 {
			continue
		}
		start := fset.Position(g.Pos()).Offset
		end := fset.Position(g.End()).Offset
		intervals = append(intervals, [2]int{start, end})
	}
	return intervals
}

// inInterval reports whether p lies inside any of the supplied intervals.
func inInterval(p int, intervals [][2]int) bool {
	for _, iv := range intervals {
		if p >= iv[0] && p < iv[1] {
			return true
		}
	}
	return false
}

// extractFromSource parses a single Go source string and returns all entity
// references found. The function is exposed independently so unit tests can
// pass synthetic source snippets. When intervals is non-nil, matches that fall
// inside a comment interval are ignored.
func extractFromSource(src string, intervals [][2]int) []EntityRef {
	var out []EntityRef
	add := func(componentSuffix, name string) {
		component, ok := componentName(componentSuffix)
		if !ok {
			return
		}
		out = append(out, EntityRef{Component: component, Name: name})
	}

	extractMatches := func(re *regexp.Regexp, isResourceBase bool) {
		for _, loc := range re.FindAllStringSubmatchIndex(src, -1) {
			// loc[0]..loc[1] is the full match; loc[2]..loc[3] is group 1, etc.
			if len(loc) < 4 {
				continue
			}
			if inInterval(loc[0], intervals) {
				continue
			}
			if isResourceBase {
				// groups: 1 = component suffix, 2 = name
				if len(loc) >= 6 {
					add(src[loc[2]:loc[3]], src[loc[4]:loc[5]])
				}
			} else {
				// groups: 1 = name (component fixed)
				add("Elasticsearch", src[loc[2]:loc[3]])
			}
		}
	}

	extractMatches(newResourceBaseRE, true)
	extractMatches(newElasticsearchResourceRE, false)
	extractMatches(newKibanaResourceRE, true)
	extractMatches(newKibanaDataSourceRE, true)

	return out
}
