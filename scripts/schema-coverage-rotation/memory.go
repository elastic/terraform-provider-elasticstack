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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	fwprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Memory holds the schema-coverage rotation memory file contents.
type Memory struct {
	Resources   map[string]*time.Time `json:"resources"`
	DataSources map[string]*time.Time `json:"data-sources"`
}

// entity represents a single provider entity with its type and name.
type entity struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

const (
	entityTypeResource   = "resource"
	entityTypeDataSource = "data source"
)

// loadMemory reads and parses the memory file at path.
func loadMemory(path string) (*Memory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read memory file: %w", err)
	}

	// Use raw map to handle null timestamps.
	var raw struct {
		Resources   map[string]any `json:"resources"`
		DataSources map[string]any `json:"data-sources"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse memory file: %w", err)
	}

	mem := &Memory{
		Resources:   make(map[string]*time.Time),
		DataSources: make(map[string]*time.Time),
	}

	for k, v := range raw.Resources {
		ts := parseTimestamp(v)
		mem.Resources[k] = ts
	}
	for k, v := range raw.DataSources {
		ts := parseTimestamp(v)
		mem.DataSources[k] = ts
	}

	return mem, nil
}

// parseTimestamp converts a raw JSON value to *time.Time (nil for null).
func parseTimestamp(v any) *time.Time {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

// saveMemory atomically writes memory to path via a temporary file rename.
func saveMemory(path string, mem *Memory) error {
	// Build raw map for serialization (null timestamps as JSON null).
	type rawMemory struct {
		Resources   map[string]any `json:"resources"`
		DataSources map[string]any `json:"data-sources"`
	}
	raw := rawMemory{
		Resources:   make(map[string]any),
		DataSources: make(map[string]any),
	}
	for k, v := range mem.Resources {
		if v == nil {
			raw.Resources[k] = nil
		} else {
			raw.Resources[k] = v.UTC().Format(time.RFC3339)
		}
	}
	for k, v := range mem.DataSources {
		if v == nil {
			raw.DataSources[k] = nil
		} else {
			raw.DataSources[k] = v.UTC().Format(time.RFC3339)
		}
	}

	data, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".schema-coverage-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

// discoverEntities builds the canonical entity inventory from provider registrations.
// It calls Resources() and DataSources() on the Plugin Framework provider, and
// reads ResourcesMap/DataSourcesMap from the Plugin SDK provider.
func discoverEntities(fwProv fwprovider.Provider, sdkResources map[string]struct{}, sdkDataSources map[string]struct{}) (resources map[string]struct{}, dataSources map[string]struct{}) {
	resources = make(map[string]struct{})
	dataSources = make(map[string]struct{})

	// Plugin Framework resources.
	ctx := context.Background()
	for _, rf := range fwProv.Resources(ctx) {
		r := rf()
		var metaResp resource.MetadataResponse
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "elasticstack"}, &metaResp)
		if metaResp.TypeName != "" {
			resources[metaResp.TypeName] = struct{}{}
		}
	}

	// Plugin Framework data sources.
	for _, dsf := range fwProv.DataSources(ctx) {
		ds := dsf()
		var metaResp datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "elasticstack"}, &metaResp)
		if metaResp.TypeName != "" {
			dataSources[metaResp.TypeName] = struct{}{}
		}
	}

	// Plugin SDK resources.
	for name := range sdkResources {
		resources[name] = struct{}{}
	}

	// Plugin SDK data sources.
	for name := range sdkDataSources {
		dataSources[name] = struct{}{}
	}

	return resources, dataSources
}

// reconcileMemory updates memory to match the canonical entity inventory.
// It adds missing entities with nil timestamps and removes stale entries.
func reconcileMemory(mem *Memory, resources, dataSources map[string]struct{}) {
	// Add missing resources.
	for name := range resources {
		if _, ok := mem.Resources[name]; !ok {
			mem.Resources[name] = nil
		}
	}
	// Remove stale resources.
	for name := range mem.Resources {
		if _, ok := resources[name]; !ok {
			delete(mem.Resources, name)
		}
	}

	// Add missing data sources.
	for name := range dataSources {
		if _, ok := mem.DataSources[name]; !ok {
			mem.DataSources[name] = nil
		}
	}
	// Remove stale data sources.
	for name := range mem.DataSources {
		if _, ok := dataSources[name]; !ok {
			delete(mem.DataSources, name)
		}
	}
}

// selectEntities selects the n oldest entities across resources and data sources.
// Null timestamps sort first; ties break by type then name (lexicographic).
func selectEntities(mem *Memory, n int) []entity {
	type candidate struct {
		entity
		ts *time.Time
	}

	var candidates []candidate
	for name, ts := range mem.Resources {
		candidates = append(candidates, candidate{entity: entity{Type: entityTypeResource, Name: name}, ts: ts})
	}
	for name, ts := range mem.DataSources {
		candidates = append(candidates, candidate{entity: entity{Type: entityTypeDataSource, Name: name}, ts: ts})
	}

	sort.Slice(candidates, func(i, j int) bool {
		ci, cj := candidates[i], candidates[j]
		// Null (never analyzed) sorts before any timestamp.
		if ci.ts == nil && cj.ts != nil {
			return true
		}
		if ci.ts != nil && cj.ts == nil {
			return false
		}
		// Both nil: break by type then name.
		if ci.ts == nil && cj.ts == nil {
			if ci.Type != cj.Type {
				return ci.Type < cj.Type
			}
			return ci.Name < cj.Name
		}
		// Both non-nil: older first, then by type then name.
		if ci.ts.Equal(*cj.ts) {
			if ci.Type != cj.Type {
				return ci.Type < cj.Type
			}
			return ci.Name < cj.Name
		}
		return ci.ts.Before(*cj.ts)
	})

	if n > len(candidates) {
		n = len(candidates)
	}
	result := make([]entity, n)
	for i := range n {
		result[i] = candidates[i].entity
	}
	return result
}
