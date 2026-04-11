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
	fmt.Fprintln(w, "  record   --memory <path> --type <t> --name <n>    Record analysis timestamp")
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

	// Bootstrap from seed if the working file does not exist.
	if _, err := os.Stat(*memPath); os.IsNotExist(err) {
		seedPath := ".github/aw/memory/schema-coverage.json"
		if err := bootstrapFromSeed(*memPath, seedPath); err != nil {
			return fmt.Errorf("bootstrap memory: %w", err)
		}
		fmt.Fprintf(stderr, "bootstrapped memory from %s\n", seedPath)
	}

	// Load current memory.
	mem, err := loadMemory(*memPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}

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
	reconcileMemory(mem, resources, dataSources)

	if err := saveMemory(*memPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	fmt.Fprintf(stderr, "prepared memory: %d resources, %d data-sources\n",
		len(mem.Resources), len(mem.DataSources))
	return nil
}

// bootstrapFromSeed copies the seed memory file to the target path,
// creating any intermediate directories.
func bootstrapFromSeed(targetPath, seedPath string) error {
	seedData, err := os.ReadFile(seedPath)
	if err != nil {
		// If the seed does not exist, start with empty memory.
		if os.IsNotExist(err) {
			empty := &Memory{
				Resources:   make(map[string]*time.Time),
				DataSources: make(map[string]*time.Time),
			}
			return saveMemory(targetPath, empty)
		}
		return fmt.Errorf("read seed: %w", err)
	}

	// Parse and re-save via saveMemory to get canonical format.
	var raw struct {
		Resources   map[string]any `json:"resources"`
		DataSources map[string]any `json:"data-sources"`
	}
	if err := json.Unmarshal(seedData, &raw); err != nil {
		return fmt.Errorf("parse seed: %w", err)
	}

	mem := &Memory{
		Resources:   make(map[string]*time.Time),
		DataSources: make(map[string]*time.Time),
	}
	for k, v := range raw.Resources {
		mem.Resources[k] = parseTimestamp(v)
	}
	for k, v := range raw.DataSources {
		mem.DataSources[k] = parseTimestamp(v)
	}

	return saveMemory(targetPath, mem)
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

// cmdRecord records the current UTC timestamp for an analyzed entity.
func cmdRecord(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to the live working memory file (required)")
	entityType := fs.String("type", "", "entity type: 'resource' or 'data source' (required)")
	entityName := fs.String("name", "", "entity name (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	if *entityType == "" {
		return errors.New("--type is required")
	}
	if *entityName == "" {
		return errors.New("--name is required")
	}
	if *entityType != entityTypeResource && *entityType != entityTypeDataSource {
		return fmt.Errorf("--type must be %q or %q", entityTypeResource, entityTypeDataSource)
	}

	mem, err := loadMemory(*memPath)
	if err != nil {
		return fmt.Errorf("load memory: %w", err)
	}

	now := time.Now().UTC()
	switch *entityType {
	case entityTypeResource:
		mem.Resources[*entityName] = &now
	case entityTypeDataSource:
		mem.DataSources[*entityName] = &now
	}

	if err := saveMemory(*memPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}

	fmt.Fprintf(stderr, "recorded %s %q at %s\n", *entityType, *entityName, now.Format(time.RFC3339))
	return nil
}
