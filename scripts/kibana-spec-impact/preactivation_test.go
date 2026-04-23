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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDerivePreActivationOutputs(t *testing.T) {
	tests := []struct {
		name         string
		report       *ImpactReport
		issueCap     int
		wantRun      bool
		wantCnt      int
		wantReason   string
		wantIssueCap int
	}{
		{
			name:         "nil report is missing",
			report:       nil,
			wantRun:      false,
			wantCnt:      0,
			wantReason:   gateReasonMissingReport,
			wantIssueCap: defaultIssueCap,
		},
		{
			name:         "high confidence impacts run agent",
			report:       &ImpactReport{HighConfidence: []ImpactedEntity{{EntityName: "entity"}}},
			wantRun:      true,
			wantCnt:      1,
			wantReason:   gateReasonHighConfidenceImpactsPresent,
			wantIssueCap: defaultIssueCap,
		},
		{
			name:         "transform hints still run agent",
			report:       &ImpactReport{TransformSchemaHints: []string{"internal/kibana/foo/transform_schema.go"}},
			wantRun:      true,
			wantCnt:      0,
			wantReason:   gateReasonTransformSchemaHintsPresent,
			wantIssueCap: defaultIssueCap,
		},
		{
			name:         "no actionable impacts skips agent",
			report:       &ImpactReport{},
			wantRun:      false,
			wantCnt:      0,
			wantReason:   gateReasonNoActionableImpacts,
			wantIssueCap: defaultIssueCap,
		},
		{
			name:         "custom issue cap is preserved",
			report:       &ImpactReport{},
			issueCap:     7,
			wantRun:      false,
			wantCnt:      0,
			wantReason:   gateReasonNoActionableImpacts,
			wantIssueCap: 7,
		},
		{
			name: "kbapi symbols changed with no high-confidence or hints still runs agent",
			report: &ImpactReport{
				ChangedKbapiSymbols: []string{"SomeAPI.Method"},
			},
			wantRun:      true,
			wantCnt:      0,
			wantReason:   gateReasonKbapiSymbolsChanged,
			wantIssueCap: defaultIssueCap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := derivePreActivationOutputs(tt.report, tt.issueCap)
			if got.ShouldRun != tt.wantRun || got.HighConfidenceCount != tt.wantCnt || got.GateReason != tt.wantReason {
				t.Fatalf("unexpected outputs: %+v", got)
			}
			if got.IssueCap != tt.wantIssueCap {
				t.Fatalf("expected issue cap %d, got %d", tt.wantIssueCap, got.IssueCap)
			}
		})
	}
}

func TestWriteReportFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "report.json")
	report := &ImpactReport{Version: 1, BaselineSHA: "b", TargetSHA: "t"}
	if err := writeReportFile(path, report); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.HasSuffix(text, "\n") {
		t.Fatal("expected trailing newline")
	}
	if !strings.Contains(text, "\"baseline_sha\": \"b\"") || !strings.Contains(text, "\"target_sha\": \"t\"") {
		t.Fatalf("unexpected report content: %s", text)
	}
}

func TestWriteReportFileError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeReportFile(filepath.Join(blocker, "report.json"), &ImpactReport{}); err == nil {
		t.Fatal("expected writeReportFile to fail")
	}
}

func TestAppendGithubOutputs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github-output.txt")
	outputs := PreActivationOutputs{
		ShouldRun:           true,
		IssueCap:            5,
		HighConfidenceCount: 2,
		GateReason:          gateReasonHighConfidenceImpactsPresent,
	}
	if err := appendGithubOutputs(path, outputs); err != nil {
		t.Fatal(err)
	}
	if err := appendGithubOutputs(path, outputs); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	want := strings.Join([]string{
		"run_agent=true",
		"issue_cap=5",
		"high_confidence_count=2",
		"gate_reason=high_confidence_impacts_present",
		"run_agent=true",
		"issue_cap=5",
		"high_confidence_count=2",
		"gate_reason=high_confidence_impacts_present",
		"",
	}, "\n")
	if text != want {
		t.Fatalf("unexpected output file:\n%s", text)
	}
}

func TestAppendGithubOutputsError(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendGithubOutputs(filepath.Join(blocker, "child"), PreActivationOutputs{}); err == nil {
		t.Fatal("expected appendGithubOutputs to fail")
	}
}

func TestEnsurePreActivationMemory(t *testing.T) {
	t.Run("bootstraps missing memory from seed", func(t *testing.T) {
		dir := t.TempDir()
		seedPath := filepath.Join(dir, "seed.json")
		memPath := filepath.Join(dir, "nested", "memory.json")
		seed := &Memory{Version: memoryVersion, LastAnalyzedTargetSHA: "seed-sha"}
		if err := saveMemory(seedPath, seed); err != nil {
			t.Fatal(err)
		}
		if err := ensurePreActivationMemory(memPath, seedPath); err != nil {
			t.Fatal(err)
		}
		mem, err := loadMemory(memPath)
		if err != nil {
			t.Fatal(err)
		}
		if mem.LastAnalyzedTargetSHA != "seed-sha" {
			t.Fatalf("expected bootstrapped memory, got %+v", mem)
		}
	})

	t.Run("leaves existing memory unchanged", func(t *testing.T) {
		dir := t.TempDir()
		seedPath := filepath.Join(dir, "seed.json")
		memPath := filepath.Join(dir, "memory.json")
		seed := &Memory{Version: memoryVersion, LastAnalyzedTargetSHA: "seed-sha"}
		if err := saveMemory(seedPath, seed); err != nil {
			t.Fatal(err)
		}
		existing := &Memory{Version: memoryVersion, LastAnalyzedTargetSHA: "existing-sha"}
		if err := saveMemory(memPath, existing); err != nil {
			t.Fatal(err)
		}
		if err := ensurePreActivationMemory(memPath, seedPath); err != nil {
			t.Fatal(err)
		}
		mem, err := loadMemory(memPath)
		if err != nil {
			t.Fatal(err)
		}
		if mem.LastAnalyzedTargetSHA != "existing-sha" {
			t.Fatalf("expected existing memory to remain unchanged, got %+v", mem)
		}
	})
}
