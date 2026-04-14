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

// Command schema-coverage-rotation provides helper commands for the
// schema-coverage-rotation workflow to manage its entity memory file.
//
// Usage:
//
//	go run ./scripts/schema-coverage-rotation <command> [flags]
//
// Commands:
//
//	prepare   Bootstrap and reconcile the memory file from provider registrations.
//	select    Select the next N entities by oldest timestamp; prints a JSON array.
//	record    Record the current UTC timestamp for an analyzed entity.
//
// All commands accept --memory <path> to specify the live working memory file.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/provider"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}

	switch args[0] {
	case "prepare":
		return cmdPrepare(args[1:], stderr)
	case "select":
		return cmdSelect(args[1:], stdout, stderr)
	case "record":
		return cmdRecord(args[1:], stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: schema-coverage-rotation <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  prepare  --memory <path>                          Bootstrap and reconcile memory")
	fmt.Fprintln(w, "  select   --memory <path> --count <n>              Select next N entities (JSON)")
	fmt.Fprintln(w, "  record   --memory <path> (--type <t> --name <n> | --entities <json>)    Record analysis timestamps")
	return errors.New("unknown or missing command")
}

// cmdPrepare bootstraps the memory file if needed and reconciles the entity inventory.
func cmdPrepare(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("prepare", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to the live working memory file (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}

	startedFromScratch := false
	seedPath := ".github/aw/memory/schema-coverage.json"
	if _, err := os.Stat(*memPath); os.IsNotExist(err) {
		startedFromScratch = true
		if err := bootstrapFromSeed(*memPath, seedPath); err != nil {
			return fmt.Errorf("bootstrap memory: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("stat memory: %w", err)
	}

	// Load current memory.
	mem, err := loadMemory(*memPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}

	migrationStats := normalizeMemoryKeys(mem)

	// Discover entities from provider registrations.
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

	resources, dataSources := discoverEntities(fwProv, sdkResources, sdkDataSources)
	stats := reconcileMemory(mem, resources, dataSources)

	if err := saveMemory(*memPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	if startedFromScratch {
		fmt.Fprintf(stderr, "bootstrapped memory from %s\n", seedPath)
		fmt.Fprintf(stderr, "prepare started from scratch at %s\n", *memPath)
	} else {
		fmt.Fprintf(stderr, "prepare re-used existing state from %s\n", *memPath)
	}
	if migrationStats.MigratedTotal() > 0 || migrationStats.CollisionTotal() > 0 {
		fmt.Fprintf(stderr, "prepare migrated legacy keys: %d resources, %d data-sources (%d resource collisions, %d data-source collisions)\n",
			migrationStats.MigratedResources, migrationStats.MigratedDataSources,
			migrationStats.MigratedResourceCollisions, migrationStats.MigratedDataSourceCollisions)
	}
	fmt.Fprintf(stderr, "prepare reconciled state: added %d, removed %d (%d resources added, %d resources removed, %d data-sources added, %d data-sources removed)\n",
		stats.AddedTotal(), stats.RemovedTotal(),
		stats.AddedResources, stats.RemovedResources,
		stats.AddedDataSources, stats.RemovedDataSources)
	fmt.Fprintf(stderr, "prepared memory: %d resources, %d data-sources\n",
		len(mem.Resources), len(mem.DataSources))
	return nil
}

// bootstrapFromSeed copies the seed memory file to the target path. If the
// seed file does not exist, it creates an empty memory file instead.
func bootstrapFromSeed(targetPath, seedPath string) error {
	var mem *Memory

	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		mem = &Memory{
			Resources:   make(map[string]*time.Time),
			DataSources: make(map[string]*time.Time),
		}
	} else if err != nil {
		return fmt.Errorf("stat seed: %w", err)
	} else {
		var err error
		mem, err = loadMemory(seedPath)
		if err != nil {
			return fmt.Errorf("load seed: %w", err)
		}
	}

	if err := saveMemory(targetPath, mem); err != nil {
		return fmt.Errorf("save bootstrapped memory: %w", err)
	}

	return nil
}

// cmdSelect selects the next N entities and emits a JSON array to stdout.
func cmdSelect(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to the live working memory file (required)")
	count := fs.Int("count", 1, "number of entities to select")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	if *count < 1 {
		return errors.New("--count must be at least 1")
	}

	mem, err := loadMemory(*memPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}

	selected := selectEntities(mem, *count)

	data, err := json.MarshalIndent(selected, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal selection: %w", err)
	}
	fmt.Fprintln(stdout, string(data))
	return nil
}

// cmdRecord records the current UTC timestamp for one or more analyzed entities
// in a single atomic memory write, preventing lost updates when multiple entities
// are recorded together.
//
// Usage (single entity):  record --memory <path> --type <t> --name <n>
// Usage (multiple entities): record --memory <path> --entities <json>
//
// --entities accepts the JSON array produced by the select command so the caller
// can pass select output directly without re-parsing it.
func cmdRecord(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to the live working memory file (required)")
	entityType := fs.String("type", "", "entity type: 'resource' or 'data source'")
	entityName := fs.String("name", "", "entity name")
	entitiesJSON := fs.String("entities", "", "JSON array of entities to record (e.g. output of 'select'); mutually exclusive with --type/--name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}

	var entities []entity
	switch {
	case *entitiesJSON != "" && (*entityType != "" || *entityName != ""):
		return errors.New("--entities and --type/--name are mutually exclusive")
	case *entitiesJSON != "":
		if err := json.Unmarshal([]byte(*entitiesJSON), &entities); err != nil {
			return fmt.Errorf("parse --entities JSON: %w", err)
		}
	default:
		// Single-entity form.
		if *entityType == "" {
			return errors.New("--type is required when --entities is not specified")
		}
		if *entityName == "" {
			return errors.New("--name is required when --entities is not specified")
		}
		entities = []entity{{Type: *entityType, Name: *entityName}}
	}

	if len(entities) == 0 {
		return errors.New("no entities to record")
	}
	for _, e := range entities {
		if e.Type != entityTypeResource && e.Type != entityTypeDataSource {
			return fmt.Errorf("invalid entity type %q: must be %q or %q", e.Type, entityTypeResource, entityTypeDataSource)
		}
		if e.Name == "" {
			return errors.New("entity name must not be empty")
		}
	}

	mem, err := loadMemory(*memPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}

	now := time.Now().UTC()
	for _, e := range entities {
		switch e.Type {
		case entityTypeResource:
			mem.Resources[e.Name] = &now
		case entityTypeDataSource:
			mem.DataSources[e.Name] = &now
		}
		fmt.Fprintf(stderr, "recorded %s %q at %s\n", e.Type, e.Name, now.Format(time.RFC3339))
	}

	if err := saveMemory(*memPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	return nil
}
