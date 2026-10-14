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
	"cmp"
	"fmt"
	"slices"
	"strings"
	"time"
)

const (
	maxTriagePerRun = 15

	gateReasonQueueReady      = "triage_queue_ready"
	gateReasonNoSlots         = "no_slots"
	gateReasonNothingToTriage = "no_untriaged_groups"
)

// QueueItem is one gap group selected for triage together with its pre-rendered issue.
type QueueItem struct {
	Key         string `json:"key"`
	Resource    string `json:"resource"`
	Fingerprint string `json:"fingerprint"`
	Size        int    `json:"size"`
	Title       string `json:"title"`
	Body        string `json:"body"`
}

// Report is report.json: the ordered triage queue plus the full diagnostics of a discovery run.
type Report struct {
	RunAgent   bool        `json:"run_agent"`
	GateReason string      `json:"gate_reason"`
	Slots      int         `json:"slots"`
	Queue      []QueueItem `json:"queue"`
	Discovery
}

// buildReport selects and orders the groups to triage. Groups covered by memory are skipped;
// the rest are ordered by resource (never-triaged first, then oldest last_triaged, then name)
// and within a resource by size (largest first, then key), and cut at maxTriagePerRun.
func buildReport(d *Discovery, mem *Memory, slots int) *Report {
	slots = max(slots, 0)
	report := &Report{
		Slots:     slots,
		Queue:     []QueueItem{},
		Discovery: *d,
	}
	if slots > 0 {
		report.Queue = triageQueue(d.Groups, mem)
	}
	switch {
	case slots == 0:
		report.GateReason = gateReasonNoSlots
	case len(report.Queue) == 0:
		report.GateReason = gateReasonNothingToTriage
	default:
		report.GateReason = gateReasonQueueReady
		report.RunAgent = true
	}
	return report
}

func triageQueue(groups []Group, mem *Memory) []QueueItem {
	pending := slices.DeleteFunc(slices.Clone(groups), mem.covers)
	slices.SortFunc(pending, func(a, b Group) int {
		if a.Resource != b.Resource {
			return compareResources(mem, a.Resource, b.Resource)
		}
		return cmp.Or(cmp.Compare(len(b.Members), len(a.Members)), strings.Compare(a.Key, b.Key))
	})
	pending = pending[:min(len(pending), maxTriagePerRun)]
	queue := make([]QueueItem, 0, len(pending))
	for _, g := range pending {
		queue = append(queue, QueueItem{
			Key:         g.Key,
			Resource:    g.Resource,
			Fingerprint: g.Fingerprint,
			Size:        len(g.Members),
			Title:       renderTitle(g),
			Body:        renderBody(g),
		})
	}
	return queue
}

func compareResources(mem *Memory, a, b string) int {
	lastA, triagedA := mem.Resources[a]
	lastB, triagedB := mem.Resources[b]
	switch {
	case triagedA != triagedB:
		if !triagedA {
			return -1
		}
		return 1
	case triagedA:
		if c := parseTriageTime(lastA).Compare(parseTriageTime(lastB)); c != 0 {
			return c
		}
	}
	return strings.Compare(a, b)
}

// parseTriageTime reads a last_triaged value: an RFC 3339 timestamp, or a bare date. An
// unreadable value sorts as the oldest, so the resource is triaged soon rather than starved.
func parseTriageTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339, time.DateOnly} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func renderDiscoverSummary(r *Report) string {
	var b strings.Builder
	b.WriteString("## Kibana API gap discovery\n\n")
	fmt.Fprintf(&b, "- run_agent: %t (%s)\n- available slots: %d\n- gap groups: %d\n- triage queue: %d\n- skipped shapes: %d\n\n",
		r.RunAgent, r.GateReason, r.Slots, len(r.Groups), len(r.Queue), len(r.SkippedShapes))
	if len(r.Queue) > 0 {
		b.WriteString("### Triage queue\n\n")
		for i, item := range r.Queue {
			fmt.Fprintf(&b, "%d. `%s` (%d fields)\n", i+1, item.Key, item.Size)
		}
		b.WriteString("\n")
	}
	if len(r.NoOperations) > 0 {
		fmt.Fprintf(&b, "Resources with no mapped operation: %s\n", strings.Join(r.NoOperations, ", "))
	}
	return b.String()
}
