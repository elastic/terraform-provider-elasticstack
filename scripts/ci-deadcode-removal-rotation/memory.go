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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type AttemptReason string

const (
	ReasonPRCreated             AttemptReason = "pr_created"
	ReasonBuildFailed           AttemptReason = "build_failed"
	ReasonTestsFailed           AttemptReason = "tests_failed"
	ReasonVerificationTimeout   AttemptReason = "verification_timeout"
	ReasonInvalidAcceptanceTest AttemptReason = "invalid_candidate_acceptance_test"
	ReasonInvalidReferences     AttemptReason = "invalid_candidate_references"
	ReasonAgentAbort            AttemptReason = "agent_abort"
	ReasonNoCandidate           AttemptReason = "no_candidate_available"
	ReasonPreActivationBlocked  AttemptReason = "preactivation_blocked"
	ReasonFmtFailed             AttemptReason = "fmt_failed"
)

var validReasons = map[AttemptReason]struct{}{
	ReasonPRCreated:             {},
	ReasonBuildFailed:           {},
	ReasonTestsFailed:           {},
	ReasonVerificationTimeout:   {},
	ReasonInvalidAcceptanceTest: {},
	ReasonInvalidReferences:     {},
	ReasonAgentAbort:            {},
	ReasonNoCandidate:           {},
	ReasonPreActivationBlocked:  {},
	ReasonFmtFailed:             {},
}

type AttemptContext struct {
	ReferenceFileCount  int      `json:"referenceFileCount,omitempty"`
	TestCleanupEligible bool     `json:"testCleanupEligible,omitempty"`
	ImpactedPackages    []string `json:"impactedPackages,omitempty"`
	BuildExitCode       int      `json:"buildExitCode,omitempty"`
	TestFailedPackages  []string `json:"testFailedPackages,omitempty"`
}

type AttemptRecord struct {
	Symbol      string         `json:"symbol"`
	Package     string         `json:"package"`
	AttemptedAt time.Time      `json:"attemptedAt"`
	Reason      AttemptReason  `json:"reason"`
	Context     AttemptContext `json:"context,omitzero"`
}

type Memory struct {
	Version      int             `json:"version"`
	CooldownDays int             `json:"cooldownDays"`
	Attempts     []AttemptRecord `json:"attempts"`
}

const maxAttempts = 500

func loadMemory(path string) (*Memory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var mem Memory
	if err := json.Unmarshal(data, &mem); err != nil {
		return nil, fmt.Errorf("parse memory: %w", err)
	}
	if mem.Version == 0 {
		mem.Version = 1
	}
	return &mem, nil
}

func saveMemory(path string, mem *Memory) error {
	data, err := json.MarshalIndent(mem, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal memory: %w", err)
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".deadcode-memory-*.json.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

func isInCooldown(mem *Memory, symbol string, now time.Time) bool {
	if mem.CooldownDays <= 0 {
		return false
	}
	cutoff := now.AddDate(0, 0, -mem.CooldownDays)
	for _, a := range slices.Backward(mem.Attempts) {
		if a.Symbol == symbol {
			return !a.AttemptedAt.Before(cutoff)
		}
	}
	return false
}

func recordAttempt(mem *Memory, symbol, pkg string, reason AttemptReason, ctx AttemptContext) {
	mem.Attempts = append(mem.Attempts, AttemptRecord{
		Symbol:      symbol,
		Package:     pkg,
		AttemptedAt: time.Now().UTC(),
		Reason:      reason,
		Context:     ctx,
	})
	trimAttempts(mem, maxAttempts)
}

func trimAttempts(mem *Memory, limit int) {
	if len(mem.Attempts) <= limit {
		return
	}
	mem.Attempts = mem.Attempts[len(mem.Attempts)-limit:]
}

func validateReason(r string) (AttemptReason, error) {
	reason := AttemptReason(r)
	if _, ok := validReasons[reason]; !ok {
		return "", fmt.Errorf("invalid reason %q", r)
	}
	return reason, nil
}

func summarize(mem *Memory, days int) string {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	counts := make(map[AttemptReason]int)
	pkgCounts := make(map[string]int)
	var recent int
	for _, a := range mem.Attempts {
		if a.AttemptedAt.Before(cutoff) {
			continue
		}
		recent++
		counts[a.Reason]++
		if a.Reason != ReasonPRCreated {
			pkgCounts[a.Package]++
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("## Dead-code removal outcomes (last %d days)", days))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("Total attempts: %d", recent))
	lines = append(lines, "")
	lines = append(lines, "| Reason | Count |")
	lines = append(lines, "|--------|-------|")
	var reasons []AttemptReason
	for r := range counts {
		reasons = append(reasons, r)
	}
	slices.SortFunc(reasons, func(a, b AttemptReason) int {
		return strings.Compare(string(a), string(b))
	})
	for _, r := range reasons {
		lines = append(lines, fmt.Sprintf("| %s | %d |", r, counts[r]))
	}
	lines = append(lines, "")

	if len(pkgCounts) > 0 {
		lines = append(lines, "Sticky packages (non-PR outcomes):")
		var pkgs []string
		for p := range pkgCounts {
			pkgs = append(pkgs, p)
		}
		slices.SortFunc(pkgs, func(a, b string) int {
			ca, cb := pkgCounts[a], pkgCounts[b]
			if ca != cb {
				// Descending by count.
				if ca > cb {
					return -1
				}
				return 1
			}
			return strings.Compare(a, b)
		})
		for _, p := range pkgs {
			lines = append(lines, fmt.Sprintf("- `%s`: %d", p, pkgCounts[p]))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}
