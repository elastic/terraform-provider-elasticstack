// Package main implements a targeted acceptance test package selector.
package main

import (
	"fmt"
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

// componentName maps the Go identifier suffix (e.g. "Kibana", "APM") to the
// string value used in Terraform type names.
func componentName(suffix string) (string, bool) {
	switch suffix {
	case "Elasticsearch":
		return "elasticsearch", true
	case "Kibana":
		return "kibana", true
	case "Fleet":
		return "fleet", true
	case "APM":
		return "apm", true
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
		for _, ent := range extractFromSource(string(data)) {
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

// extractFromSource parses a single Go source string and returns all entity
// references found. The function is exposed independently so unit tests can
// pass synthetic source snippets.
func extractFromSource(src string) []EntityRef {
	var out []EntityRef
	add := func(componentSuffix, name string) {
		component, ok := componentName(componentSuffix)
		if !ok {
			return
		}
		out = append(out, EntityRef{Component: component, Name: name})
	}

	for _, m := range newResourceBaseRE.FindAllStringSubmatch(src, -1) {
		if len(m) >= 3 {
			add(m[1], m[2])
		}
	}

	for _, m := range newElasticsearchResourceRE.FindAllStringSubmatch(src, -1) {
		if len(m) >= 2 {
			add("Elasticsearch", m[1])
		}
	}

	for _, m := range newKibanaResourceRE.FindAllStringSubmatch(src, -1) {
		if len(m) >= 3 {
			add(m[1], m[2])
		}
	}

	for _, m := range newKibanaDataSourceRE.FindAllStringSubmatch(src, -1) {
		if len(m) >= 3 {
			add(m[1], m[2])
		}
	}

	return out
}
