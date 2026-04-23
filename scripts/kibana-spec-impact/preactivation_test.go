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
		name       string
		report     *ImpactReport
		wantRun    bool
		wantCnt    int
		wantReason string
	}{
		{
			name:       "high confidence impacts run agent",
			report:     &ImpactReport{HighConfidence: []ImpactedEntity{{EntityName: "entity"}}},
			wantRun:    true,
			wantCnt:    1,
			wantReason: "high_confidence_impacts_present",
		},
		{
			name:       "transform hints still run agent",
			report:     &ImpactReport{TransformSchemaHints: []string{"internal/kibana/foo/transform_schema.go"}},
			wantRun:    true,
			wantCnt:    0,
			wantReason: "transform_schema_hints_present",
		},
		{
			name:       "no actionable impacts skips agent",
			report:     &ImpactReport{},
			wantRun:    false,
			wantCnt:    0,
			wantReason: "no_actionable_impacts",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := derivePreActivationOutputs(tt.report, 0)
			if got.ShouldRun != tt.wantRun || got.HighConfidenceCount != tt.wantCnt || got.GateReason != tt.wantReason {
				t.Fatalf("unexpected outputs: %+v", got)
			}
			if got.IssueCap != defaultIssueCap {
				t.Fatalf("expected default issue cap %d, got %d", defaultIssueCap, got.IssueCap)
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

func TestAppendGithubOutputs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "github-output.txt")
	outputs := PreActivationOutputs{
		ShouldRun:           true,
		IssueCap:            5,
		HighConfidenceCount: 2,
		GateReason:          "high_confidence_impacts_present",
	}
	if err := appendGithubOutputs(path, outputs); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"should_run=true",
		"issue_cap=5",
		"high_confidence_count=2",
		"gate_reason=high_confidence_impacts_present",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in output file: %s", want, text)
		}
	}
}
