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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func gapGroup(resource, field string, size int, fingerprint string) Group {
	g := Group{
		Key:         resource + "#/components/schemas/W/properties/" + field,
		Resource:    resource,
		Field:       field,
		Pointer:     "/components/schemas/W/properties/" + field,
		Fingerprint: fingerprint,
	}
	for i := range size {
		g.Members = append(g.Members, Member{Field: fmt.Sprintf("%s.m%d", field, i), Pointer: g.Pointer + fmt.Sprintf("/m%d", i)})
	}
	return g
}

func queueKeys(r *Report) []string {
	keys := make([]string, 0, len(r.Queue))
	for _, item := range r.Queue {
		keys = append(keys, item.Key)
	}
	return keys
}

func TestBuildReportOrdersNeverTriagedResourcesThenOldestThenNameThenSizeThenKey(t *testing.T) {
	d := &Discovery{Groups: []Group{
		gapGroup("r_old", "a", 1, "f"),
		gapGroup("r_new", "a", 1, "f"),
		gapGroup("r_never_b", "a", 1, "f"),
		gapGroup("r_never_a", "big", 3, "f"),
		gapGroup("r_never_a", "small", 1, "f"),
		gapGroup("r_never_a", "also_small", 1, "f"),
	}}
	mem := &Memory{Resources: map[string]string{"r_old": "2026-01-01", "r_new": "2026-03-01"}}

	r := buildReport(d, mem, 3)

	assert.Equal(t, []string{
		"r_never_a#/components/schemas/W/properties/big",
		"r_never_a#/components/schemas/W/properties/also_small",
		"r_never_a#/components/schemas/W/properties/small",
		"r_never_b#/components/schemas/W/properties/a",
		"r_old#/components/schemas/W/properties/a",
		"r_new#/components/schemas/W/properties/a",
	}, queueKeys(r))
}

func TestBuildReportBreaksLastTriagedTiesByResourceName(t *testing.T) {
	d := &Discovery{Groups: []Group{gapGroup("r_b", "a", 1, "f"), gapGroup("r_a", "a", 1, "f")}}
	mem := &Memory{Resources: map[string]string{"r_a": "2026-01-01", "r_b": "2026-01-01"}}

	assert.Equal(t, []string{
		"r_a#/components/schemas/W/properties/a",
		"r_b#/components/schemas/W/properties/a",
	}, queueKeys(buildReport(d, mem, 3)))
}

func TestBuildReportCapsTheQueueAtFifteenGroups(t *testing.T) {
	d := &Discovery{}
	for i := range 20 {
		d.Groups = append(d.Groups, gapGroup("r", fmt.Sprintf("f%02d", i), 1, "f"))
	}

	r := buildReport(d, &Memory{}, 3)

	assert.Len(t, r.Queue, 15)
	assert.Equal(t, 15, maxTriagePerRun)
	assert.Len(t, r.Groups, 20, "diagnostics keep every group")
}

func TestBuildReportSkipsGroupsCoveredByAnEqualFingerprint(t *testing.T) {
	covered := gapGroup("r", "covered", 1, "same")
	expired := gapGroup("r", "expired", 1, "new")
	fresh := gapGroup("r", "fresh", 1, "f")
	mem := &Memory{Groups: map[string]Record{
		covered.Key: {Fingerprint: "same", Verdict: verdictSuppress, Reason: "x"},
		expired.Key: {Fingerprint: "old", Verdict: verdictSuppress, Reason: "x"},
	}}

	r := buildReport(&Discovery{Groups: []Group{covered, expired, fresh}}, mem, 3)

	assert.Equal(t, []string{expired.Key, fresh.Key}, queueKeys(r))
}

func TestBuildReportIgnoresRecordsForGroupsNoLongerDiscovered(t *testing.T) {
	mem := &Memory{Groups: map[string]Record{"gone#/x": {Fingerprint: "f", Verdict: verdictFile}}}

	r := buildReport(&Discovery{Groups: []Group{gapGroup("r", "a", 1, "f")}}, mem, 3)

	assert.Len(t, r.Queue, 1)
}

func TestBuildReportGatesTheAgentOnSlotsAndQueue(t *testing.T) {
	d := &Discovery{Groups: []Group{gapGroup("r", "a", 1, "f")}}

	ready := buildReport(d, &Memory{}, 1)
	assert.True(t, ready.RunAgent)
	assert.Equal(t, gateReasonQueueReady, ready.GateReason)
	assert.Equal(t, 1, ready.Slots)

	noSlots := buildReport(d, &Memory{}, 0)
	assert.False(t, noSlots.RunAgent)
	assert.Equal(t, gateReasonNoSlots, noSlots.GateReason)
	assert.Empty(t, noSlots.Queue, "no slots means no triage")

	negative := buildReport(d, &Memory{}, -2)
	assert.False(t, negative.RunAgent)
	assert.Equal(t, 0, negative.Slots)

	empty := buildReport(&Discovery{}, &Memory{}, 3)
	assert.False(t, empty.RunAgent)
	assert.Equal(t, gateReasonNothingToTriage, empty.GateReason)
}

func TestBuildReportCarriesRenderedIssuesAndDiagnostics(t *testing.T) {
	d := &Discovery{
		Groups:        []Group{gapGroup("r", "a", 1, "fp1")},
		SkippedShapes: []SkippedShape{{Resource: "r", Pointer: "/p", Shape: "not"}},
		Operations:    map[string]int{"r": 2},
		NoOperations:  []string{"es"},
	}

	r := buildReport(d, &Memory{}, 1)

	require.Len(t, r.Queue, 1)
	assert.Equal(t, "r: a", r.Queue[0].Title)
	assert.Contains(t, r.Queue[0].Body, "<!-- kibana-api-gap key="+r.Queue[0].Key+" fp=fp1 -->")
	assert.Equal(t, "r", r.Queue[0].Resource)
	assert.Equal(t, "fp1", r.Queue[0].Fingerprint)
	assert.Equal(t, d.SkippedShapes, r.SkippedShapes)
	assert.Equal(t, d.Operations, r.Operations)
	assert.Equal(t, d.NoOperations, r.NoOperations)
}

func TestBuildReportIsDeterministic(t *testing.T) {
	d := &Discovery{Groups: []Group{gapGroup("r", "a", 2, "f"), gapGroup("s", "b", 2, "f")}}
	first := buildReport(d, &Memory{}, 2)
	second := buildReport(d, &Memory{}, 2)
	assert.Equal(t, first, second)
}

func TestRenderSummaryMentionsQueueGateAndDiagnostics(t *testing.T) {
	d := &Discovery{
		Groups:       []Group{gapGroup("r", "a", 1, "f")},
		Operations:   map[string]int{"r": 2},
		NoOperations: []string{"es"},
	}

	summary := renderDiscoverSummary(buildReport(d, &Memory{}, 1))

	assert.Contains(t, summary, "run_agent")
	assert.Contains(t, summary, gateReasonQueueReady)
	assert.Contains(t, summary, "r#/components/schemas/W/properties/a")
	assert.Contains(t, summary, "es")
}

func TestFinishDiscoverWritesReportAndSummary(t *testing.T) {
	dir := t.TempDir()
	d := &Discovery{Groups: []Group{gapGroup("r", "a", 1, "f")}}
	opts := discoverOptions{slots: 1, memoryPath: filepath.Join(dir, "state.json"), reportPath: filepath.Join(dir, "report.json"), summaryPath: filepath.Join(dir, "summary.md")}

	require.NoError(t, finishDiscover(d, opts, io.Discard))

	data, err := os.ReadFile(opts.reportPath)
	require.NoError(t, err)
	var got Report
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, got.RunAgent)
	assert.Len(t, got.Queue, 1)
	summary, err := os.ReadFile(opts.summaryPath)
	require.NoError(t, err)
	assert.Contains(t, string(summary), "Kibana API gap discovery")
}

func TestFinishDiscoverSkipsGroupsCoveredByTheMemoryFile(t *testing.T) {
	dir := t.TempDir()
	g := gapGroup("r", "a", 1, "f")
	opts := discoverOptions{slots: 1, memoryPath: filepath.Join(dir, "state.json"), reportPath: filepath.Join(dir, "report.json")}
	require.NoError(t, saveMemory(opts.memoryPath, &Memory{Groups: map[string]Record{g.Key: {Fingerprint: "f", Verdict: verdictSuppress, Reason: "x"}}}))

	require.NoError(t, finishDiscover(&Discovery{Groups: []Group{g}}, opts, io.Discard))

	data, err := os.ReadFile(opts.reportPath)
	require.NoError(t, err)
	var got Report
	require.NoError(t, json.Unmarshal(data, &got))
	assert.False(t, got.RunAgent)
	assert.Equal(t, gateReasonNothingToTriage, got.GateReason)
}

func TestFinishDiscoverIsByteIdenticalAcrossRuns(t *testing.T) {
	dir := t.TempDir()
	d := &Discovery{Groups: []Group{gapGroup("r", "a", 2, "f"), gapGroup("s", "b", 1, "f")}}
	read := func(name string) []byte {
		opts := discoverOptions{slots: 2, memoryPath: filepath.Join(dir, "none.json"), reportPath: filepath.Join(dir, name)}
		require.NoError(t, finishDiscover(d, opts, io.Discard))
		data, err := os.ReadFile(opts.reportPath)
		require.NoError(t, err)
		return data
	}
	assert.Equal(t, read("1.json"), read("2.json"))
}
