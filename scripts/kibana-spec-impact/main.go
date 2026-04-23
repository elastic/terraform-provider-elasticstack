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

// Command kibana-spec-impact provides deterministic helpers for the Kibana OpenAPI
// spec-impact workflow: entity inventory, kbapi diffing, impact reporting, and memory.
//
// Set TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true to include experimental Kibana entities
// in inventory/report (matches CI).
//
// Usage:
//
//	go run ./scripts/kibana-spec-impact <command> [flags]
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
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
	case "inventory":
		return cmdInventory(args[1:], stdout, stderr)
	case "report":
		return cmdReport(args[1:], stdout, stderr)
	case "resolve-baseline":
		return cmdResolveBaseline(args[1:], stdout, stderr)
	case "pre-activation":
		return cmdPreActivation(args[1:], stderr)
	case "memory-bootstrap":
		return cmdMemoryBootstrap(args[1:], stderr)
	case "memory-record-from-report":
		return cmdMemoryRecordFromReport(args[1:], stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: kibana-spec-impact <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  inventory                    Print JSON entity inventory (stdout)")
	fmt.Fprintln(w, "  report                       Emit JSON impact report for baseline..target")
	fmt.Fprintln(w, "  resolve-baseline             Print resolved baseline SHA for the target revision")
	fmt.Fprintln(w, "  pre-activation               Bootstrap memory if needed, write report, and emit workflow outputs")
	fmt.Fprintln(w, "  memory-bootstrap             Copy seed memory to --memory if missing")
	fmt.Fprintln(w, "  memory-record-from-report    Advance baseline; --issued required when report has high_confidence_impacts")
	return errors.New("unknown or missing command")
}

func cmdInventory(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("inventory", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	entities := discoverKibanaEntities()
	data, err := json.MarshalIndent(entities, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = stdout.Write(data)
	return err
}

func cmdReport(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo", ".", "repository root")
	memPath := fs.String("memory", "", "optional path to live memory file for duplicate suppression")
	target := fs.String("target", "HEAD", "git revision for the analysis target")
	baselineOverride := fs.String("baseline", "", "optional baseline revision (skips resolve-baseline)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var mem *Memory
	if *memPath != "" {
		var err error
		mem, err = loadMemory(*memPath)
		if err != nil {
			return err
		}
	}

	baseline := *baselineOverride
	if baseline == "" {
		var err error
		baseline, err = resolveBaseline(*repoRoot, mem, *target)
		if err != nil {
			return fmt.Errorf("resolve baseline: %w", err)
		}
	}

	report, err := buildImpactReport(*repoRoot, mem, baseline, *target)
	if err != nil {
		return err
	}
	data, err := encodeReport(report)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = stdout.Write(data)
	return err
}

func resolveBaseline(repoRoot string, mem *Memory, target string) (string, error) {
	if mem != nil && mem.LastAnalyzedTargetSHA != "" {
		return gitRevParse(repoRoot, mem.LastAnalyzedTargetSHA)
	}
	t, err := gitRevParse(repoRoot, target)
	if err != nil {
		return "", err
	}
	return gitRevParse(repoRoot, t+"~1")
}

func cmdResolveBaseline(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("resolve-baseline", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo", ".", "repository root")
	memPath := fs.String("memory", "", "optional path to memory file")
	target := fs.String("target", "HEAD", "git revision for the analysis target")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var mem *Memory
	if *memPath != "" {
		var err error
		mem, err = loadMemory(*memPath)
		if err != nil {
			return err
		}
	}
	baseline, err := resolveBaseline(*repoRoot, mem, *target)
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, baseline)
	return nil
}

func cmdPreActivation(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("pre-activation", flag.ContinueOnError)
	fs.SetOutput(stderr)
	repoRoot := fs.String("repo", ".", "repository root")
	memPath := fs.String("memory", "", "path to live memory file (required)")
	seedPath := fs.String("seed", ".github/aw/memory/kibana-spec-impact.json", "seed memory path")
	target := fs.String("target", "HEAD", "git revision for the analysis target")
	baselineOverride := fs.String("baseline", "", "optional baseline revision (skips resolve-baseline)")
	reportPath := fs.String("report-path", "kibana-spec-impact-report.json", "path to write the report JSON")
	issueCap := fs.Int("issue-cap", defaultIssueCap, "maximum issues the workflow may create")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	if err := ensurePreActivationMemory(*memPath, *seedPath); err != nil {
		return err
	}
	mem, err := loadMemory(*memPath)
	if err != nil {
		return err
	}
	baseline := *baselineOverride
	if baseline == "" {
		baseline, err = resolveBaseline(*repoRoot, mem, *target)
		if err != nil {
			return fmt.Errorf("resolve baseline: %w", err)
		}
	}
	report, err := buildImpactReport(*repoRoot, mem, baseline, *target)
	if err != nil {
		return err
	}
	if err := writeReportFile(*reportPath, report); err != nil {
		return err
	}
	outputs := derivePreActivationOutputs(report, *issueCap)
	if outputFile := os.Getenv("GITHUB_OUTPUT"); outputFile != "" {
		if err := appendGithubOutputs(outputFile, outputs); err != nil {
			return fmt.Errorf("write github outputs: %w", err)
		}
	}
	fmt.Fprintf(
		stderr,
		"wrote report to %s; run_agent=%t issue_cap=%d high_confidence_count=%d gate_reason=%s\n",
		*reportPath,
		outputs.ShouldRun,
		outputs.IssueCap,
		outputs.HighConfidenceCount,
		outputs.GateReason,
	)
	return nil
}

func ensurePreActivationMemory(memPath, seedPath string) error {
	if _, err := os.Stat(memPath); os.IsNotExist(err) {
		if err := bootstrapMemoryFromSeed(memPath, seedPath); err != nil {
			return fmt.Errorf("bootstrap memory: %w", err)
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func cmdMemoryBootstrap(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("memory-bootstrap", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to live memory file (required)")
	seedPath := fs.String("seed", ".github/aw/memory/kibana-spec-impact.json", "seed memory path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	if _, err := os.Stat(*memPath); err == nil {
		fmt.Fprintf(stderr, "memory already exists at %s\n", *memPath)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return bootstrapMemoryFromSeed(*memPath, *seedPath)
}

func cmdMemoryRecordFromReport(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("memory-record-from-report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to live memory file (required)")
	reportPath := fs.String("report", "", "path to report JSON (required)")
	issuedPath := fs.String("issued", "", "JSON array of entity names that received an issue this run (use [] if none). Required when the report includes high_confidence_impacts.")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" || *reportPath == "" {
		return errors.New("--memory and --report are required")
	}
	raw, err := os.ReadFile(*reportPath)
	if err != nil {
		return err
	}
	var report ImpactReport
	if err := json.Unmarshal(raw, &report); err != nil {
		return err
	}
	if err := validateMemoryRecordCLI(&report, *issuedPath); err != nil {
		return err
	}
	mem, err := loadMemory(*memPath)
	if err != nil {
		return err
	}

	var issuedNames []string
	if *issuedPath != "" {
		issuedRaw, err := os.ReadFile(*issuedPath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(issuedRaw, &issuedNames); err != nil {
			return fmt.Errorf("parse --issued: %w", err)
		}
	}

	recorded, err := recordIssuedFingerprints(mem, &report, issuedNames)
	if err != nil {
		return err
	}
	advanceMemoryBaseline(mem, report.TargetSHA)

	if err := saveMemory(*memPath, mem); err != nil {
		return err
	}
	fmt.Fprintf(stderr, "recorded %d fingerprint(s); advanced last_analyzed_target_sha to %s\n", recorded, mem.LastAnalyzedTargetSHA)
	return nil
}

func validateMemoryRecordCLI(report *ImpactReport, issuedPath string) error {
	if len(report.HighConfidence) > 0 && issuedPath == "" {
		return errors.New("report contains high_confidence_impacts: --issued path to a JSON array is required (use a file containing [] when no issues were created)")
	}
	return nil
}
