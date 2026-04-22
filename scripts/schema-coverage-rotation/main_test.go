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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/provider"
)

// writeMemoryFile writes a memory file with the given JSON content.
func writeMemoryFile(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "schema-coverage.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write memory file: %v", err)
	}
	return path
}

func currentProviderInventory() (map[string]struct{}, map[string]struct{}) {
	fwProv := provider.NewFrameworkProvider("schema-coverage-rotation")
	sdkProv := provider.New("schema-coverage-rotation")

	sdkResources := make(map[string]struct{})
	for name := range sdkProv.ResourcesMap {
		sdkResources[name] = struct{}{}
	}

	sdkDataSources := make(map[string]struct{})
	for name := range sdkProv.DataSourcesMap {
		sdkDataSources[name] = struct{}{}
	}

	return discoverEntities(fwProv, sdkResources, sdkDataSources)
}

// TestCmdPrepareMissingMemoryFlag checks that --memory is required.
func TestCmdPrepareMissingMemoryFlag(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	err := cmdPrepare([]string{}, &stderr)
	if err == nil {
		t.Error("expected error for missing --memory flag")
	}
}

// TestCmdPrepareBootstrapsFromSeed verifies that a missing working file is
// bootstrapped from the repo seed, then reconciled against current registrations.
func TestCmdPrepareBootstrapsFromSeed(t *testing.T) {
	dir := t.TempDir()

	seedDir := filepath.Join(dir, ".github", "aw", "memory")
	if err := os.MkdirAll(seedDir, 0o755); err != nil {
		t.Fatalf("mkdir seed dir: %v", err)
	}

	seedContent := `{
		"resources": {"elasticsearch_index_template": "2021-01-01T00:00:00Z"},
		"data-sources": {"elasticstack_seed_only": null}
	}`
	if err := os.WriteFile(filepath.Join(seedDir, "schema-coverage.json"), []byte(seedContent), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	t.Chdir(dir)

	memPath := filepath.Join(dir, "working", "schema-coverage.json")
	var stderr bytes.Buffer
	if err := cmdPrepare([]string{"--memory", memPath}, &stderr); err != nil {
		t.Fatalf("cmdPrepare: %v\nstderr: %s", err, stderr.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "bootstrapped memory from .github/aw/memory/schema-coverage.json") {
		t.Fatalf("expected bootstrap log, got stderr: %s", output)
	}
	if !strings.Contains(output, "prepare started from scratch at "+memPath) {
		t.Fatalf("expected scratch log, got stderr: %s", output)
	}
	if !strings.Contains(output, "prepare migrated legacy keys: 1 resources, 0 data-sources (0 resource collisions, 0 data-source collisions)") {
		t.Fatalf("expected migration log, got stderr: %s", output)
	}

	mem, err := loadMemory(memPath)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}
	if _, ok := mem.Resources["elasticstack_elasticsearch_index_template"]; !ok {
		t.Fatal("expected bootstrapped resource key to be normalized")
	}
	if _, ok := mem.DataSources["elasticstack_seed_only"]; ok {
		t.Fatal("expected unregistered seed-only data source to be removed during reconcile")
	}
}

func TestCmdPrepareMissingMemoryFileWithoutSeedStartsEmpty(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	memPath := filepath.Join(dir, "working", "schema-coverage.json")
	var stderr bytes.Buffer
	if err := cmdPrepare([]string{"--memory", memPath}, &stderr); err != nil {
		t.Fatalf("cmdPrepare: %v\nstderr: %s", err, stderr.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "bootstrapped memory from .github/aw/memory/schema-coverage.json") {
		t.Fatalf("expected bootstrap log, got stderr: %s", output)
	}
	if !strings.Contains(output, "prepare started from scratch at "+memPath) {
		t.Fatalf("expected scratch log, got stderr: %s", output)
	}
}

// TestCmdPrepareReconciles verifies that new/stale entities are handled and logged.
func TestCmdPrepareReconciles(t *testing.T) {
	dir := t.TempDir()
	resources, dataSources := currentProviderInventory()

	staleResource := "elasticstack_test_only_stale_resource"
	staleDataSource := "elasticstack_test_only_stale_data_source"
	if _, ok := resources[staleResource]; ok {
		t.Fatalf("stale resource name unexpectedly registered: %s", staleResource)
	}
	if _, ok := dataSources[staleDataSource]; ok {
		t.Fatalf("stale data source name unexpectedly registered: %s", staleDataSource)
	}

	memPath := writeMemoryFile(t, dir, fmt.Sprintf(`{
		"resources": {"%s": null},
		"data-sources": {"%s": null}
	}`, staleResource, staleDataSource))

	var stderr bytes.Buffer
	if err := cmdPrepare([]string{"--memory", memPath}, &stderr); err != nil {
		t.Fatalf("cmdPrepare: %v\nstderr: %s", err, stderr.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "prepare re-used existing state from "+memPath) {
		t.Fatalf("expected reuse log, got stderr: %s", output)
	}
	expectedLog := fmt.Sprintf(
		"prepare reconciled state: added %d, removed %d (%d resources added, %d resources removed, %d data-sources added, %d data-sources removed)",
		len(resources)+len(dataSources), 2,
		len(resources), 1,
		len(dataSources), 1,
	)
	if !strings.Contains(output, expectedLog) {
		t.Fatalf("expected reconcile log %q, got stderr: %s", expectedLog, output)
	}

	mem, err := loadMemory(memPath)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	// After prepare the memory should contain provider-registered entities.
	if len(mem.Resources) == 0 && len(mem.DataSources) == 0 {
		t.Error("expected at least some registered entities after prepare")
	}
	if _, ok := mem.Resources[staleResource]; ok {
		t.Errorf("expected stale resource %q to be removed", staleResource)
	}
	if _, ok := mem.DataSources[staleDataSource]; ok {
		t.Errorf("expected stale data source %q to be removed", staleDataSource)
	}
}

func TestCmdPrepareMigratesLegacyKeys(t *testing.T) {
	t.Parallel()

	resources, dataSources := currentProviderInventory()
	resourceName := "elasticstack_elasticsearch_index_template"
	dataSourceName := "elasticstack_elasticsearch_index_template"
	if _, ok := resources[resourceName]; !ok {
		t.Fatalf("expected resource %q to be registered", resourceName)
	}
	if _, ok := dataSources[dataSourceName]; !ok {
		t.Fatalf("expected data source %q to be registered", dataSourceName)
	}

	dir := t.TempDir()
	memPath := writeMemoryFile(t, dir, `{
		"resources": {"elasticsearch_index_template": "2021-01-01T00:00:00Z"},
		"data-sources": {"elasticsearch_index_template": "2022-02-02T00:00:00Z"}
	}`)

	var stderr bytes.Buffer
	if err := cmdPrepare([]string{"--memory", memPath}, &stderr); err != nil {
		t.Fatalf("cmdPrepare: %v\nstderr: %s", err, stderr.String())
	}

	output := stderr.String()
	if !strings.Contains(output, "prepare migrated legacy keys: 1 resources, 1 data-sources (0 resource collisions, 0 data-source collisions)") {
		t.Fatalf("expected migration log, got stderr: %s", output)
	}

	mem, err := loadMemory(memPath)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	if _, ok := mem.Resources["elasticsearch_index_template"]; ok {
		t.Fatal("expected legacy resource key to be removed")
	}
	if _, ok := mem.DataSources["elasticsearch_index_template"]; ok {
		t.Fatal("expected legacy data-source key to be removed")
	}

	resourceTS := mem.Resources[resourceName]
	if resourceTS == nil || resourceTS.Format(time.RFC3339) != "2021-01-01T00:00:00Z" {
		t.Fatalf("expected migrated resource timestamp to be preserved, got %v", resourceTS)
	}

	dataSourceTS := mem.DataSources[dataSourceName]
	if dataSourceTS == nil || dataSourceTS.Format(time.RFC3339) != "2022-02-02T00:00:00Z" {
		t.Fatalf("expected migrated data-source timestamp to be preserved, got %v", dataSourceTS)
	}
}

// TestCmdSelectBasic verifies that the select command emits valid JSON.
func TestCmdSelectBasic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{
		"resources": {
			"elasticstack_a": null,
			"elasticstack_b": "2021-01-01T00:00:00Z"
		},
		"data-sources": {}
	}`)

	var stdout, stderr bytes.Buffer
	if err := cmdSelect([]string{"--memory", path, "--count", "1"}, &stdout, &stderr); err != nil {
		t.Fatalf("cmdSelect: %v\nstderr: %s", err, stderr.String())
	}

	var result []entity
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal result: %v\nstdout: %s", err, stdout.String())
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result))
	}
	// Null timestamp first.
	if result[0].Name != "elasticstack_a" {
		t.Errorf("expected elasticstack_a (null ts first), got %q", result[0].Name)
	}
	if result[0].Type != entityTypeResource {
		t.Errorf("expected resource type, got %q", result[0].Type)
	}
}

// TestCmdSelectMissingMemoryFlag checks that --memory is required.
func TestCmdSelectMissingMemoryFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := cmdSelect([]string{}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for missing --memory flag")
	}
}

// TestCmdSelectInvalidCount checks that --count < 1 fails.
func TestCmdSelectInvalidCount(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{"resources":{},"data-sources":{}}`)
	var stdout, stderr bytes.Buffer
	err := cmdSelect([]string{"--memory", path, "--count", "0"}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for count=0")
	}
}

// TestCmdRecordPersistsTimestamp verifies that record updates the memory file.
func TestCmdRecordPersistsTimestamp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{
		"resources": {"elasticstack_target": null},
		"data-sources": {}
	}`)

	// Truncate to seconds since timestamps are stored in RFC3339 (second precision).
	before := time.Now().UTC().Truncate(time.Second)

	var stderr bytes.Buffer
	if err := cmdRecord([]string{
		"--memory", path,
		"--type", "resource",
		"--name", "elasticstack_target",
	}, &stderr); err != nil {
		t.Fatalf("cmdRecord: %v\nstderr: %s", err, stderr.String())
	}

	after := time.Now().UTC().Add(time.Second)

	mem, err := loadMemory(path)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	ts, ok := mem.Resources["elasticstack_target"]
	if !ok {
		t.Fatal("entity not found in memory after record")
	}
	if ts == nil {
		t.Fatal("timestamp should not be nil after record")
	}
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", *ts, before, after)
	}
}

// TestCmdRecordDataSource verifies that a data source timestamp is recorded.
func TestCmdRecordDataSource(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{
		"resources": {},
		"data-sources": {"elasticstack_my_ds": null}
	}`)

	var stderr bytes.Buffer
	if err := cmdRecord([]string{
		"--memory", path,
		"--type", "data source",
		"--name", "elasticstack_my_ds",
	}, &stderr); err != nil {
		t.Fatalf("cmdRecord: %v", err)
	}

	mem, err := loadMemory(path)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	ts := mem.DataSources["elasticstack_my_ds"]
	if ts == nil {
		t.Error("data source timestamp should not be nil after record")
	}
}

// TestCmdRecordMissingMemoryFlag checks that --memory is required.
func TestCmdRecordMissingMemoryFlag(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	err := cmdRecord([]string{"--type", "resource", "--name", "elasticstack_x"}, &stderr)
	if err == nil {
		t.Error("expected error for missing --memory flag")
	}
}

// TestCmdRecordInvalidType checks that an unknown type is rejected.
func TestCmdRecordInvalidType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{"resources":{},"data-sources":{}}`)
	var stderr bytes.Buffer
	err := cmdRecord([]string{
		"--memory", path,
		"--type", "unknown",
		"--name", "elasticstack_x",
	}, &stderr)
	if err == nil {
		t.Error("expected error for invalid entity type")
	}
}

// TestCmdRecordEntitiesJSON verifies that --entities records multiple entities atomically.
func TestCmdRecordEntitiesJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{
		"resources": {"elasticstack_a": null, "elasticstack_b": null},
		"data-sources": {"elasticstack_ds": null}
	}`)

	entitiesJSON := `[{"type":"resource","name":"elasticstack_a"},{"type":"data source","name":"elasticstack_ds"}]`
	before := time.Now().UTC().Truncate(time.Second)

	var stderr bytes.Buffer
	if err := cmdRecord([]string{"--memory", path, "--entities", entitiesJSON}, &stderr); err != nil {
		t.Fatalf("cmdRecord: %v\nstderr: %s", err, stderr.String())
	}

	after := time.Now().UTC().Add(time.Second)

	mem, err := loadMemory(path)
	if err != nil {
		t.Fatalf("loadMemory: %v", err)
	}

	for name, ts := range map[string]*time.Time{
		"elasticstack_a (resource)":     mem.Resources["elasticstack_a"],
		"elasticstack_ds (data source)": mem.DataSources["elasticstack_ds"],
	} {
		if ts == nil {
			t.Errorf("%s: timestamp should not be nil after record", name)
			continue
		}
		if ts.Before(before) || ts.After(after) {
			t.Errorf("%s: timestamp %v not in range [%v, %v]", name, *ts, before, after)
		}
	}
	// elasticstack_b was not in the entities list and should remain nil.
	if mem.Resources["elasticstack_b"] != nil {
		t.Errorf("elasticstack_b: should remain nil, got %v", *mem.Resources["elasticstack_b"])
	}
}

// TestCmdRecordEntitiesInvalidJSON checks that malformed --entities JSON is rejected.
func TestCmdRecordEntitiesInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{"resources":{},"data-sources":{}}`)
	var stderr bytes.Buffer
	if err := cmdRecord([]string{"--memory", path, "--entities", "not-json"}, &stderr); err == nil {
		t.Error("expected error for invalid --entities JSON")
	}
}

// TestCmdRecordEntitiesMutuallyExclusive verifies --entities and --type/--name cannot be combined.
func TestCmdRecordEntitiesMutuallyExclusive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{"resources":{"elasticstack_x":null},"data-sources":{}}`)
	var stderr bytes.Buffer
	err := cmdRecord([]string{
		"--memory", path,
		"--entities", `[{"type":"resource","name":"elasticstack_x"}]`,
		"--type", "resource",
		"--name", "elasticstack_x",
	}, &stderr)
	if err == nil {
		t.Error("expected error when --entities and --type/--name are combined")
	}
}

// TestRunUnknownCommand checks that an unknown command is rejected.
func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{"bogus"}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

// TestRunNoArgs checks that missing command is rejected.
func TestRunNoArgs(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{}, &stdout, &stderr)
	if err == nil {
		t.Error("expected error for no args")
	}
}

// TestCmdRecordAtomic verifies no tmp files remain after a record.
func TestCmdRecordAtomic(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := writeMemoryFile(t, dir, `{"resources":{"elasticstack_x":null},"data-sources":{}}`)

	var stderr bytes.Buffer
	if err := cmdRecord([]string{
		"--memory", path,
		"--type", "resource",
		"--name", "elasticstack_x",
	}, &stderr); err != nil {
		t.Fatalf("cmdRecord: %v", err)
	}

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".schema-coverage-") {
			t.Errorf("unexpected tmp file left: %s", e.Name())
		}
	}
}
