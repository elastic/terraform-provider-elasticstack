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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	ts2020 = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	ts2021 = time.Date(2021, 6, 15, 12, 0, 0, 0, time.UTC)
	ts2022 = time.Date(2022, 3, 10, 8, 0, 0, 0, time.UTC)
)

// TestLoadAndSaveMemory round-trips a memory file through save then load.
func TestLoadAndSaveMemory(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_a": new(ts2020),
			"elasticstack_b": nil,
		},
		DataSources: map[string]*time.Time{
			"elasticstack_ds_a": new(ts2021),
			"elasticstack_ds_b": nil,
		},
	}

	path := filepath.Join(t.TempDir(), "schema-coverage.json")
	if err := saveMemory(path, mem); err != nil {
		t.Fatalf("saveMemory: %v", err)
	}

	got, err := loadMemory(path)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	// Check resources.
	for name, want := range mem.Resources {
		gotTS, ok := got.Resources[name]
		if !ok {
			t.Errorf("missing resource %q", name)
			continue
		}
		if want == nil {
			if gotTS != nil {
				t.Errorf("resource %q: want nil, got %v", name, *gotTS)
			}
		} else {
			if gotTS == nil {
				t.Errorf("resource %q: want %v, got nil", name, *want)
			} else if !want.Equal(*gotTS) {
				t.Errorf("resource %q: want %v, got %v", name, *want, *gotTS)
			}
		}
	}

	// Check data sources.
	for name, want := range mem.DataSources {
		gotTS, ok := got.DataSources[name]
		if !ok {
			t.Errorf("missing data-source %q", name)
			continue
		}
		if want == nil {
			if gotTS != nil {
				t.Errorf("data-source %q: want nil, got %v", name, *gotTS)
			}
		} else {
			if gotTS == nil {
				t.Errorf("data-source %q: want %v, got nil", name, *want)
			} else if !want.Equal(*gotTS) {
				t.Errorf("data-source %q: want %v, got %v", name, *want, *gotTS)
			}
		}
	}
}

// TestSaveMemoryAtomic verifies that saveMemory does not leave tmp files
// behind and the final file is valid JSON.
func TestSaveMemoryAtomic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "schema-coverage.json")
	mem := &Memory{
		Resources:   map[string]*time.Time{"elasticstack_x": nil},
		DataSources: map[string]*time.Time{},
	}
	if err := saveMemory(path, mem); err != nil {
		t.Fatalf("saveMemory: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var check map[string]any
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, data)
	}

	// No tmp files should remain.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "schema-coverage.json" {
			t.Errorf("unexpected file left in dir: %s", e.Name())
		}
	}
}

// TestReconcileMemoryAddsNew verifies new entities are added with nil timestamps.
func TestReconcileMemoryAddsNew(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources:   map[string]*time.Time{},
		DataSources: map[string]*time.Time{},
	}
	resources := map[string]struct{}{
		"elasticstack_new_resource": {},
	}
	dataSources := map[string]struct{}{
		"elasticstack_new_ds": {},
	}

	reconcileMemory(mem, resources, dataSources)

	if ts, ok := mem.Resources["elasticstack_new_resource"]; !ok {
		t.Error("expected new resource to be added")
	} else if ts != nil {
		t.Errorf("expected nil timestamp for new resource, got %v", *ts)
	}

	if ts, ok := mem.DataSources["elasticstack_new_ds"]; !ok {
		t.Error("expected new data source to be added")
	} else if ts != nil {
		t.Errorf("expected nil timestamp for new data source, got %v", *ts)
	}
}

// TestReconcileMemoryPreservesTimestamp verifies existing timestamps are preserved.
func TestReconcileMemoryPreservesTimestamp(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_existing": new(ts2021),
		},
		DataSources: map[string]*time.Time{},
	}
	resources := map[string]struct{}{
		"elasticstack_existing": {},
	}

	reconcileMemory(mem, resources, map[string]struct{}{})

	ts, ok := mem.Resources["elasticstack_existing"]
	if !ok {
		t.Fatal("existing resource removed unexpectedly")
	}
	if ts == nil || !ts.Equal(ts2021) {
		t.Errorf("expected preserved timestamp %v, got %v", ts2021, ts)
	}
}

// TestReconcileMemoryRemovesStale verifies entities no longer registered are removed.
func TestReconcileMemoryRemovesStale(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_stale": new(ts2020),
			"elasticstack_kept":  new(ts2021),
		},
		DataSources: map[string]*time.Time{
			"elasticstack_stale_ds": nil,
		},
	}
	resources := map[string]struct{}{
		"elasticstack_kept": {},
	}
	dataSources := map[string]struct{}{}

	reconcileMemory(mem, resources, dataSources)

	if _, ok := mem.Resources["elasticstack_stale"]; ok {
		t.Error("stale resource should have been removed")
	}
	if _, ok := mem.Resources["elasticstack_kept"]; !ok {
		t.Error("kept resource should not have been removed")
	}
	if _, ok := mem.DataSources["elasticstack_stale_ds"]; ok {
		t.Error("stale data source should have been removed")
	}
}

// TestSelectEntitiesNullFirst verifies entities with nil timestamps come first.
func TestSelectEntitiesNullFirst(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_analyzed": new(ts2020),
			"elasticstack_never":    nil,
		},
		DataSources: map[string]*time.Time{},
	}

	selected := selectEntities(mem, 1)
	if len(selected) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(selected))
	}
	if selected[0].Name != "elasticstack_never" {
		t.Errorf("expected null-timestamp entity first, got %q", selected[0].Name)
	}
}

// TestSelectEntitiesOldestFirst verifies older timestamps are selected before newer.
func TestSelectEntitiesOldestFirst(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_old":    new(ts2020),
			"elasticstack_recent": new(ts2022),
		},
		DataSources: map[string]*time.Time{},
	}

	selected := selectEntities(mem, 1)
	if len(selected) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(selected))
	}
	if selected[0].Name != "elasticstack_old" {
		t.Errorf("expected oldest entity, got %q", selected[0].Name)
	}
}

// TestSelectEntitiesTieBreaking verifies ties break by type then name.
func TestSelectEntitiesTieBreaking(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_b": nil,
			"elasticstack_a": nil,
		},
		DataSources: map[string]*time.Time{
			"elasticstack_c": nil,
		},
	}

	// "data source" < "resource" lexicographically, so DS should come first.
	selected := selectEntities(mem, 3)
	if len(selected) != 3 {
		t.Fatalf("expected 3 entities, got %d", len(selected))
	}
	if selected[0].Type != entityTypeDataSource {
		t.Errorf("expected data source first (type tie-break), got %q", selected[0].Type)
	}
	// Resources sorted by name: a before b.
	if selected[1].Name != "elasticstack_a" {
		t.Errorf("expected elasticstack_a second, got %q", selected[1].Name)
	}
	if selected[2].Name != "elasticstack_b" {
		t.Errorf("expected elasticstack_b third, got %q", selected[2].Name)
	}
}

// TestSelectEntitiesCount verifies that fewer entities than n are handled gracefully.
func TestSelectEntitiesCount(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources:   map[string]*time.Time{"elasticstack_only": nil},
		DataSources: map[string]*time.Time{},
	}

	selected := selectEntities(mem, 5)
	if len(selected) != 1 {
		t.Errorf("expected 1 entity (all available), got %d", len(selected))
	}
}

// TestSelectEntitiesEntityType verifies the entity type field is set correctly.
func TestSelectEntitiesEntityType(t *testing.T) {
	t.Parallel()

	mem := &Memory{
		Resources: map[string]*time.Time{
			"elasticstack_res": new(ts2021),
		},
		DataSources: map[string]*time.Time{
			"elasticstack_ds": new(ts2020),
		},
	}

	selected := selectEntities(mem, 2)
	if len(selected) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(selected))
	}

	// DS has older timestamp so should be first.
	if selected[0].Type != entityTypeDataSource {
		t.Errorf("expected data source first, got %q", selected[0].Type)
	}
	if selected[1].Type != entityTypeResource {
		t.Errorf("expected resource second, got %q", selected[1].Type)
	}
}
