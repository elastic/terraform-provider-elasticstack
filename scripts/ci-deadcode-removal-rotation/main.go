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

// Command ci-deadcode-removal-rotation provides deterministic helpers for the
// dead-code removal rotation workflow.
//
// Usage:
//
//	go run ./scripts/ci-deadcode-removal-rotation <command> [flags]
//
// Commands:
//
//	select    Run dual deadcode scans, intersect, filter cooldown, and select one candidate.
//	record    Record an attempt in cooldown memory with a deterministic reason code.
//	summarize Print a compact summary of recent attempt outcomes.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usageError(stderr)
	}
	switch args[0] {
	case "select":
		return cmdSelect(args[1:], stdout, stderr)
	case "record":
		return cmdRecord(args[1:], stderr)
	case "summarize":
		return cmdSummarize(args[1:], stdout, stderr)
	default:
		return usageError(stderr)
	}
}

func usageError(w io.Writer) error {
	fmt.Fprintln(w, "Usage: ci-deadcode-removal-rotation <command> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  select    --memory <path> [--cooldown-days <n>] [--module-path <path>]")
	fmt.Fprintln(w, "  record    --memory <path> --symbol <s> --package <p> --reason <r> [--context <json>]")
	fmt.Fprintln(w, "  summarize --memory <path> [--days <n>]")
	return errors.New("unknown or missing command")
}

func cmdSelect(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to cooldown memory file (required)")
	cooldownDays := fs.Int("cooldown-days", 14, "cooldown window in days")
	modulePath := fs.String("module-path", "github.com/elastic/terraform-provider-elasticstack", "Go module path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}

	fmt.Fprintf(stderr, "running deadcode without tests...\n")
	withoutTests, err := runDeadcode(false)
	if err != nil {
		return fmt.Errorf("deadcode ./...: %w", err)
	}
	fmt.Fprintf(stderr, "found %d candidates without tests\n", len(withoutTests))

	fmt.Fprintf(stderr, "running deadcode with tests...\n")
	withTests, err := runDeadcode(true)
	if err != nil {
		return fmt.Errorf("deadcode -test ./...: %w", err)
	}
	fmt.Fprintf(stderr, "found %d candidates with tests\n", len(withTests))

	for i := range withoutTests {
		withoutTests[i].packagePath = derivePackagePath(withoutTests[i].file, *modulePath)
	}
	for i := range withTests {
		withTests[i].packagePath = derivePackagePath(withTests[i].file, *modulePath)
	}

	intersected := intersectCandidates(withoutTests, withTests)
	fmt.Fprintf(stderr, "intersection size: %d\n", len(intersected))

	mem, err := loadMemory(*memPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load memory: %w", err)
		}
		mem = &Memory{
			Version:      1,
			CooldownDays: *cooldownDays,
			Attempts:     []AttemptRecord{},
		}
	}
	mem.CooldownDays = *cooldownDays

	now := time.Now().UTC()
	total := len(intersected)
	inCooldown := 0
	for _, c := range intersected {
		if isInCooldown(mem, c.key(), now) {
			inCooldown++
		}
	}

	chosen := selectOne(intersected, mem, now)

	result := Candidate{Found: false}
	if chosen != nil {
		result.Found = true
		result.Symbol = chosen.key()
		result.SymbolName = chosen.symbol
		result.Package = chosen.packagePath
		result.File = chosen.file
		result.Line = chosen.line
		result.Column = chosen.column

		refFiles, err := runGoplsReferences(chosen.file, chosen.line, chosen.column)
		if err != nil {
			return fmt.Errorf("gopls references: %w", err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get working directory: %w", err)
		}
		relRefFiles := make([]string, 0, len(refFiles))
		for _, rf := range refFiles {
			rel, err := relativePath(cwd, rf)
			if err != nil {
				return fmt.Errorf("resolve reference path %s: %w", rf, err)
			}
			relRefFiles = append(relRefFiles, rel)
		}
		result.ReferenceFiles = relRefFiles

		eligible, testFile := classifyReferences(relRefFiles)
		result.CompanionTestCleanupEligible = eligible
		result.CompanionTestFile = testFile
		result.ImpactedPackages = impactedPackages(*chosen, testFile)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Fprintln(stdout, string(data))
	fmt.Fprintf(stderr, "total=%d in-cooldown=%d eligible=%d selected=%v\n", total, inCooldown, total-inCooldown, result.Found)
	return nil
}

func cmdRecord(args []string, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to memory file (required)")
	symbol := fs.String("symbol", "", "candidate symbol (required)")
	pkg := fs.String("package", "", "candidate package (required)")
	reasonStr := fs.String("reason", "", "attempt reason code (required)")
	contextJSON := fs.String("context", "", "JSON context object")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	if *symbol == "" {
		return errors.New("--symbol is required")
	}
	if *pkg == "" {
		return errors.New("--package is required")
	}
	if *reasonStr == "" {
		return errors.New("--reason is required")
	}
	reason, err := validateReason(*reasonStr)
	if err != nil {
		return err
	}

	var ctx AttemptContext
	if *contextJSON != "" {
		if err := json.Unmarshal([]byte(*contextJSON), &ctx); err != nil {
			return fmt.Errorf("parse context JSON: %w", err)
		}
	}

	mem, err := loadMemory(*memPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("load memory: %w", err)
		}
		mem = &Memory{Version: 1, CooldownDays: 14}
	}

	recordAttempt(mem, *symbol, *pkg, reason, ctx)

	if err := saveMemory(*memPath, mem); err != nil {
		return fmt.Errorf("save memory: %w", err)
	}
	fmt.Fprintf(stderr, "recorded %s (%s) as %s\n", *symbol, *pkg, reason)
	return nil
}

func cmdSummarize(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("summarize", flag.ContinueOnError)
	fs.SetOutput(stderr)
	memPath := fs.String("memory", "", "path to memory file (required)")
	days := fs.Int("days", 30, "summary window in days")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *memPath == "" {
		return errors.New("--memory is required")
	}
	mem, err := loadMemory(*memPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(stdout, "## Dead-code removal outcomes")
			fmt.Fprintln(stdout, "")
			fmt.Fprintln(stdout, "No memory file found.")
			return nil
		}
		return fmt.Errorf("load memory: %w", err)
	}
	summary := summarize(mem, *days)
	fmt.Fprintln(stdout, summary)
	return nil
}
