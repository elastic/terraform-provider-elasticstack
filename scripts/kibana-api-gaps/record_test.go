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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

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
		{Key: "s#/c", Verdict: verdictSuppress, Reason: "internal"},
	}, testTime)

	require.NoError(t, err)
	assert.Equal(t, Record{Fingerprint: "fa", Verdict: verdictFile, Reason: "useful", Date: "2026-10-13", Issue: "r#/a"}, mem.Groups["r#/a"])
	assert.Equal(t, Record{Fingerprint: "fb", Verdict: verdictSuppress, Reason: "server-managed counter", Date: "2026-10-13"}, mem.Groups["r#/b"])
	assert.Contains(t, mem.Groups, "r#/old", "unrelated records are kept")
	assert.Equal(t, map[string]string{"r": "2026-10-13T12:00:00Z", "s": "2026-10-13T12:00:00Z"}, mem.Resources)
}

func TestApplyVerdictsReplacesAnExpiredRecord(t *testing.T) {
	mem := newMemory()
	mem.Groups["r#/a"] = Record{Fingerprint: "stale", Verdict: verdictSuppress, Reason: "old", Date: "2026-01-01"}

	require.NoError(t, applyVerdicts(recordTestReport(1), mem, []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "now useful"}}, testTime))

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

			err := applyVerdicts(recordTestReport(tc.slots), mem, tc.verdicts, testTime)

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
	}, testTime)
	require.NoError(t, err)
}

func TestApplyVerdictsIsIdempotent(t *testing.T) {
	verdicts := []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "x"}, {Key: "r#/b", Verdict: verdictSuppress, Reason: "y"}, {Key: "s#/c", Verdict: verdictSuppress, Reason: "z"}}
	once, twice := newMemory(), newMemory()

	require.NoError(t, applyVerdicts(recordTestReport(2), once, verdicts, testTime))
	require.NoError(t, applyVerdicts(recordTestReport(2), twice, verdicts, testTime))
	require.NoError(t, applyVerdicts(recordTestReport(2), twice, verdicts, testTime))

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
		{Key: "s#/c", Verdict: verdictSuppress, Reason: "internal"},
	})
	memory := filepath.Join(dir, "memory", "state.json")
	summary := filepath.Join(dir, "summary.md")

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", memory, "--time", "2026-10-13T12:00:00Z", "--summary", summary,
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

func TestRecordCommandSucceedsWhenTheStepSummaryIsUnwritable(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(1))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "useful"}})
	memory := filepath.Join(dir, "state.json")
	var stdout, stderr bytes.Buffer

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", memory, "--time", "2026-10-13T12:00:00Z",
		"--summary", filepath.Join(dir, "missing-dir", "step_summary"),
	}, &stdout, &stderr)

	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "step summary")
	assert.FileExists(t, memory)
}

func TestRecordCommandLeavesMemoryUntouchedWhenVerdictsAreRejected(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(1))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []Verdict{{Key: "bogus", Verdict: verdictFile, Reason: "x"}})
	memory := filepath.Join(dir, "state.json")

	err := cmdRecord([]string{
		"--report", filepath.Join(dir, "report.json"),
		"--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", memory, "--time", "2026-10-13T12:00:00Z",
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
		"--memory", filepath.Join(dir, "state.json"), "--time", "2026-10-13T12:00:00Z",
	}, os.Stdout, os.Stderr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "field")
}

func TestRecordCommandRequiresReportVerdictsAndMemoryPaths(t *testing.T) {
	require.Error(t, cmdRecord(nil, os.Stdout, os.Stderr))
}

func fourQueueReport(slots int) *Report {
	r := &Report{Slots: slots}
	for _, k := range []string{"r#/a", "r#/b", "r#/c", "r#/d"} {
		r.Queue = append(r.Queue, QueueItem{Key: k, Resource: "r", Fingerprint: "f" + k})
	}
	return r
}

func TestApplyVerdictsStopsAfterTheSlotsAreFilledWithFileVerdicts(t *testing.T) {
	mem := newMemory()
	err := applyVerdicts(fourQueueReport(2), mem, []Verdict{
		{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"},
		{Key: "r#/b", Verdict: verdictFile, Reason: "x"},
		{Key: "r#/c", Verdict: verdictFile, Reason: "x"},
	}, testTime)
	require.NoError(t, err, "suppress does not consume a slot, so triage stops after the third group")
	assert.Len(t, mem.Groups, 3)

	mem = newMemory()
	err = applyVerdicts(fourQueueReport(2), mem, []Verdict{
		{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"},
		{Key: "r#/b", Verdict: verdictFile, Reason: "x"},
		{Key: "r#/c", Verdict: verdictFile, Reason: "x"},
		{Key: "r#/d", Verdict: verdictSuppress, Reason: "x"},
	}, testTime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slots are filled")
	assert.Equal(t, newMemory(), mem)
}

func TestApplyVerdictsRequiresTheTriageOrder(t *testing.T) {
	skipFirst := []Verdict{{Key: "r#/b", Verdict: verdictSuppress, Reason: "x"}}
	swapped := []Verdict{{Key: "r#/b", Verdict: verdictSuppress, Reason: "x"}, {Key: "r#/a", Verdict: verdictSuppress, Reason: "x"}}
	for name, verdicts := range map[string][]Verdict{"skips the first group": skipFirst, "swapped": swapped} {
		t.Run(name, func(t *testing.T) {
			mem := newMemory()
			err := applyVerdicts(fourQueueReport(2), mem, verdicts, testTime)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "order")
			assert.Equal(t, newMemory(), mem)
		})
	}
	prefix := []Verdict{{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"}, {Key: "r#/b", Verdict: verdictFile, Reason: "x"}, {Key: "r#/c", Verdict: verdictFile, Reason: "x"}}
	require.NoError(t, applyVerdicts(fourQueueReport(2), newMemory(), prefix, testTime), "a prefix that fills the slots is accepted")
}

func TestApplyVerdictsRejectsAShortPrefixThatNeitherFillsSlotsNorCoversTheQueue(t *testing.T) {
	short := []Verdict{{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"}, {Key: "r#/b", Verdict: verdictFile, Reason: "x"}}
	mem := newMemory()
	err := applyVerdicts(fourQueueReport(2), mem, short, testTime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stopped early")
	assert.Equal(t, newMemory(), mem)

	allSuppressShort := []Verdict{{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"}, {Key: "r#/b", Verdict: verdictSuppress, Reason: "x"}}
	require.Error(t, applyVerdicts(fourQueueReport(2), newMemory(), allSuppressShort, testTime))
}

func TestApplyVerdictsAcceptsFewerFilesThanSlotsWhenTheWholeQueueIsCovered(t *testing.T) {
	all := []Verdict{
		{Key: "r#/a", Verdict: verdictSuppress, Reason: "x"},
		{Key: "r#/b", Verdict: verdictSuppress, Reason: "x"},
		{Key: "r#/c", Verdict: verdictFile, Reason: "x"},
		{Key: "r#/d", Verdict: verdictSuppress, Reason: "x"},
	}
	mem := newMemory()
	require.NoError(t, applyVerdicts(fourQueueReport(2), mem, all, testTime))
	assert.Len(t, mem.Groups, 4)
}

func TestApplyVerdictsTrimsKeyAndVerdictAndReason(t *testing.T) {
	mem := newMemory()
	require.NoError(t, applyVerdicts(fourQueueReport(1), mem, []Verdict{{Key: " r#/a\n", Verdict: " file ", Reason: "  because\n"}}, testTime))
	assert.Equal(t, Record{Fingerprint: "fr#/a", Verdict: verdictFile, Reason: "because", Date: "2026-10-13", Issue: "r#/a"}, mem.Groups["r#/a"])
}

func TestApplyVerdictsRecordsTheTimestampPerResourceForSameDayRotation(t *testing.T) {
	mem := newMemory()
	later := testTime.Add(3 * time.Hour)
	suppress := func(key string) Verdict { return Verdict{Key: key, Verdict: verdictSuppress, Reason: "x"} }
	require.NoError(t, applyVerdicts(recordTestReport(3), mem, []Verdict{suppress("r#/a"), suppress("r#/b"), suppress("s#/c")}, later))
	assert.Equal(t, "2026-10-13T15:00:00Z", mem.Resources["r"])
}

func TestRecordCommandRejectsAnInvalidTime(t *testing.T) {
	dir := t.TempDir()
	writeJSON(t, filepath.Join(dir, "report.json"), recordTestReport(1))
	writeJSON(t, filepath.Join(dir, "verdicts.json"), []Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "x"}})
	err := cmdRecord([]string{"--report", filepath.Join(dir, "report.json"), "--verdicts", filepath.Join(dir, "verdicts.json"),
		"--memory", filepath.Join(dir, "state.json"), "--time", "yesterday"}, io.Discard, io.Discard)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "time")
}

func TestReadJSONRejectsTrailingDataAndEmptyVerdictLists(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}
	report := write("report.json", mustMarshal(t, recordTestReport(1)))
	for name, content := range map[string]string{
		"trailing": `[{"key":"r#/a","verdict":"file","reason":"x"}] []`,
		"null":     `null`,
		"empty":    `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			err := cmdRecord([]string{"--report", report, "--verdicts", write(name+".json", content), "--memory", filepath.Join(dir, "state.json")}, io.Discard, io.Discard)
			require.Error(t, err)
			assert.NoFileExists(t, filepath.Join(dir, "state.json"))
		})
	}
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return string(data)
}

func TestRenderRecordSummarySanitisesReasons(t *testing.T) {
	long := strings.Repeat("word ", 200)
	summary := renderRecordSummary([]Verdict{{Key: "r#/a", Verdict: verdictFile, Reason: "line one\n\n## injected\t heading"}, {Key: "r#/b", Verdict: verdictSuppress, Reason: long}}, testTime)

	assert.NotContains(t, summary, "line one\n")
	assert.Contains(t, summary, "line one ## injected heading")
	assert.Less(t, len(summary), 600)
	assert.Contains(t, summary, "...")
}
