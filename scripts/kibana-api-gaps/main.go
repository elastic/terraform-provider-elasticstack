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

// Command kibana-api-gaps finds Kibana API request fields that Terraform resources never use.
//
// Usage:
//
//	go run ./scripts/kibana-api-gaps discover [--repo .] [--slots 3] [--memory state.json] [--report report.json] [--resource elasticstack_kibana_slo]
//	go run ./scripts/kibana-api-gaps record --report report.json --verdicts verdicts.json --memory state.json
//	go run ./scripts/kibana-api-gaps render-issue --report report.json --key <key> --reason <reason>
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	kbapiPkgSuffix = "/generated/kbapi"
	defaultSlots   = 3
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "discover":
			return cmdDiscover(args[1:], stdout, stderr)
		case "record":
			return cmdRecord(args[1:], stdout, stderr)
		case "render-issue":
			return cmdRenderIssue(args[1:], stdout, stderr)
		}
	}
	fmt.Fprintln(stderr, "Usage: kibana-api-gaps discover|record|render-issue [flags]")
	return errors.New("unknown or missing command")
}

type stringList []string

func (s *stringList) String() string     { return fmt.Sprint([]string(*s)) }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

type discoverOptions struct {
	slots       int
	memoryPath  string
	reportPath  string
	summaryPath string
}

func cmdDiscover(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repo := fs.String("repo", ".", "repository root")
	var only stringList
	fs.Var(&only, "resource", "restrict discovery to this resource type name (repeatable)")
	var opts discoverOptions
	fs.IntVar(&opts.slots, "slots", defaultSlots, "available issue slots; zero disables triage")
	fs.StringVar(&opts.memoryPath, "memory", "state.json", "memory state.json; an absent file is an empty memory")
	fs.StringVar(&opts.reportPath, "report", "report.json", "where to write report.json")
	fs.StringVar(&opts.summaryPath, "summary", os.Getenv("GITHUB_STEP_SUMMARY"), "file to append a step summary to")
	if err := fs.Parse(args); err != nil {
		return err
	}
	log := newPhaseLog(stderr)
	result, err := discoverRepository(*repo, only, log)
	if err != nil {
		return err
	}
	if err := finishDiscover(result, opts, stdout); err != nil {
		return err
	}
	log.total()
	return nil
}

// finishDiscover turns a discovery result into report.json, a step summary, and a one-line status.
func finishDiscover(result *Discovery, opts discoverOptions, stdout io.Writer) error {
	mem, err := loadMemory(opts.memoryPath)
	if err != nil {
		return err
	}
	report := buildReport(result, mem, opts.slots)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(opts.reportPath, append(data, '\n')); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "wrote report to %s; run_agent=%t gate_reason=%s queue=%d groups=%d\n",
		opts.reportPath, report.RunAgent, report.GateReason, len(report.Queue), len(report.Groups))
	return appendSummary(opts.summaryPath, renderDiscoverSummary(report))
}

// discoverRepository runs discovery over the provider and generated/kbapi/oas-filtered.yaml
// of the repository at root, optionally restricted to the named resources.
func discoverRepository(root string, only []string, log *phaseLog) (*Discovery, error) {
	resources, err := selectResources(discoverResources(), only)
	if err != nil {
		return nil, err
	}
	log.begin("oas")
	doc, err := loadOAS(filepath.Join(root, "generated/kbapi/oas-filtered.yaml"))
	if err != nil {
		return nil, err
	}
	log.end("oas")
	log.begin("load")
	mod, err := loadModule(root, "./...")
	if err != nil {
		return nil, err
	}
	log.end("load")
	log.begin("index")
	mod.moduleIndex()
	log.end("index")
	return discoverLoaded(mod, doc, resources, log)
}

func discoverLoaded(mod *loadedModule, doc *oasDoc, resources []Resource, log *phaseLog) (*Discovery, error) {
	kbapiPath, err := findPackageBySuffix(mod, kbapiPkgSuffix)
	if err != nil {
		return nil, err
	}
	return discover(discoverConfig{module: mod, kbapiPath: kbapiPath, oas: doc, resources: resources, log: log})
}

// selectResources restricts resources to the named ones; an unknown name is an error so a
// typo cannot silently produce an empty report.
func selectResources(all []Resource, only []string) ([]Resource, error) {
	if len(only) == 0 {
		return all, nil
	}
	for _, name := range only {
		if !slices.ContainsFunc(all, func(r Resource) bool { return r.Name == name }) {
			return nil, fmt.Errorf("unknown resource %q", name)
		}
	}
	return slices.DeleteFunc(slices.Clone(all), func(r Resource) bool { return !slices.Contains(only, r.Name) }), nil
}

func findPackageBySuffix(mod *loadedModule, suffix string) (string, error) {
	var matches []string
	for path := range mod.byPath {
		if strings.HasSuffix(path, suffix) {
			matches = append(matches, path)
		}
	}
	if len(matches) != 1 {
		slices.Sort(matches)
		return "", fmt.Errorf("expected exactly one package ending in %s, found %v", suffix, matches)
	}
	return matches[0], nil
}
