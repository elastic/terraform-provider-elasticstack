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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func recordTestReport(slots int) *Report {
	return &Report{Slots: slots, Queue: []QueueItem{
		{Key: "r#/a", Resource: "r", Fingerprint: "fa"},
		{Key: "r#/b", Resource: "r", Fingerprint: "fb"},
		{Key: "s#/c", Resource: "s", Fingerprint: "fc"},
	}}
}

func TestApplyVerdictsMergesFileAndSuppressRecords(t *testing.T) {
	mem := newMemory()
	mem.Groups["r#/old"] = Record{Fingerprint: "x", Verdict: verdictFile}

	err := applyVerdicts(recordTestReport(2), mem, []Verdict{
		{Key: "r#/a", Verdict: verdictFile, Reason: "useful"},
		{Key: "r#/b", Verdict: verdictSuppress, Reason: "server-managed counter"},
	}, "2026-10-13")

	require.NoError(t, err)
	assert.Equal(t, Record{Fingerprint: "fa", Verdict: verdictFile, Reason: "useful", Date: "2026-10-13", Issue: "r#/a"}, mem.Groups["r#/a"])
	assert.Equal(t, Record{Fingerprint: "fb", Verdict: verdictSuppress, Reason: "server-managed counter", Date: "2026-10-13"}, mem.Groups["r#/b"])
	assert.Contains(t, mem.Groups, "r#/old", "unrelated records are kept")
	assert.Equal(t, map[string]string{"r": "2026-10-13"}, mem.Resources)
}

func TestApplyVerdictsReplacesAnExpiredRecord(t *testing.T) {
	mem := newMemory()
	mem.Groups["r#/a"] = Record{Fingerprint: "stale", Verdict: verdictSuppress, Reason: "old", Date: "2026-01-01"}

	require.NoError(t, applyVerdicts(recordTestReport(1), mem, []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "now useful"}}, "2026-10-13"))

	assert.Equal(t, "fa", mem.Groups["r#/a"].Fingerprint)
	assert.Equal(t, verdictFile, mem.Groups["r#/a"].Verdict)
}

func TestApplyVerdictsRejectsInvalidInputWithoutChangingMemory(t *testing.T) {
	file := func(key string) Verdict { return Verdict{Key: key, Verdict: verdictFile, Reason: "r"} }
	tests := map[string]struct {
		slots    int
		verdicts []Verdict
		contains string
	}{
		"unknown key":          {2, []Verdict{file("r#/zzz")}, "r#/zzz"},
		"too many file":        {1, []Verdict{file("r#/a"), file("r#/b")}, "slots"},
		"empty reason":         {2, []Verdict{{Key: "r#/a", Verdict: verdictSuppress, Reason: ""}}, "reason"},
		"blank reason":         {2, []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "  \n"}}, "reason"},
		"unknown verdict kind": {2, []Verdict{{Key: "r#/a", Verdict: "maybe", Reason: "r"}}, "maybe"},
		"duplicate key":        {3, []Verdict{file("r#/a"), file("r#/a")}, "duplicate"},
		"valid then invalid":   {2, []Verdict{file("r#/a"), file("r#/zzz")}, "r#/zzz"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mem := newMemory()

			err := applyVerdicts(recordTestReport(tc.slots), mem, tc.verdicts, "2026-10-13")

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.contains)
			assert.Equal(t, newMemory(), mem)
		})
	}
}

func TestApplyVerdictsAllowsUnlimitedSuppressVerdicts(t *testing.T) {
	err := applyVerdicts(recordTestReport(1), newMemory(), []Verdict{
		{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"},
		{Key: "r#/b", Verdict: verdictSuppress, Reason: "x"},
		{Key: "s#/c", Verdict: verdictFile, Reason: "x"},
	}, "2026-10-13")
	require.NoError(t, err)
}

func TestApplyVerdictsIsIdempotent(t *testing.T) {
	verdicts := []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "x"}, {Key: "r#/b", Verdict: verdictSuppress, Reason: "y"}}
	once, twice := newMemory(), newMemory()

	require.NoError(t, applyVerdicts(recordTestReport(2), once, verdicts, "2026-10-13"))
	require.NoError(t, applyVerdicts(recordTestReport(2), twice, verdicts, "2026-10-13"))
	require.NoError(t, applyVerdicts(recordTestReport(2), twice, verdicts, "2026-10-13"))

	assert.Equal(t, once, twice)
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

func TestRecordCommandCreatesMemoryAndWritesSummary(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(2))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []Verdict{
		{Key: "r#/a", Verdict: verdictFile, Reason: "useful"},
		{Key: "r#/b", Verdict: verdictSuppress, Reason: "server-managed counter"},
	})
	memory := filepath.Join(dir, "memory", "state.json")
	summary := filepath.Join(dir, "summary.md")

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", memory, "--date", "2026-10-13", "--summary", summary,
	}, os.Stdout, os.Stderr)

	require.NoError(t, err)
	mem, err := loadMemory(memory)
	require.NoError(t, err)
	assert.Equal(t, "server-managed counter", mem.Groups["r#/b"].Reason)
	text, err := os.ReadFile(summary)
	require.NoError(t, err)
	assert.Contains(t, string(text), "r#/a")
	assert.Contains(t, string(text), "server-managed counter")
}

func TestRecordCommandLeavesMemoryUntouchedWhenVerdictsAreRejected(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(1))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []Verdict{{Key: "bogus", Verdict: verdictFile, Reason: "x"}})
	memory := filepath.Join(dir, "state.json")

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", memory, "--date", "2026-10-13",
	}, os.Stdout, os.Stderr)

	require.Error(t, err)
	assert.NoFileExists(t, memory)
}

func TestRecordCommandRejectsUnknownVerdictFields(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(1))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []map[string]string{{"key": "r#/a", "verdict": "file", "reason": "x", "field": "invented"}})

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", filepath.Join(dir, "state.json"), "--date", "2026-10-13",
	}, os.Stdout, os.Stderr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "field")
}

func TestRecordCommandRequiresReportVerdictsAndMemoryPaths(t *testing.T) {
	require.Error(t, cmdRecord(nil, os.Stdout, os.Stderr))
}
