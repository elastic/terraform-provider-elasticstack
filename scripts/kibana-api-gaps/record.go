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
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Verdict is the triage outcome for one queued gap group.
type Verdict struct {
	Key     string `json:"key"`
	Verdict string `json:"verdict"`
	Reason  string `json:"reason"`
}

// applyVerdicts validates verdicts against the report's queue and merges them into mem. The
// verdicts must follow the triage order: a prefix of the queue, stopping once the slots are
// filled with `file` verdicts. mem is only modified when every verdict is valid.
func applyVerdicts(report *Report, mem *Memory, verdicts []Verdict, when time.Time) error {
	verdicts = normalise(verdicts)
	queued := make(map[string]QueueItem, len(report.Queue))
	for _, item := range report.Queue {
		queued[item.Key] = item
	}
	seen := make(map[string]bool, len(verdicts))
	files := 0
	for i, v := range verdicts {
		if _, ok := queued[v.Key]; !ok {
			return fmt.Errorf("verdict for unknown key %q: it is not in the triage queue", v.Key)
		}
		if seen[v.Key] {
			return fmt.Errorf("duplicate verdict for key %q", v.Key)
		}
		seen[v.Key] = true
		if v.Key != report.Queue[i].Key {
			return fmt.Errorf("verdict %d is for %q but the triage order requires %q", i+1, v.Key, report.Queue[i].Key)
		}
		if files >= report.Slots {
			return fmt.Errorf("verdict for %q comes after the slots are filled; triage must stop", v.Key)
		}
		if v.Verdict != verdictFile && v.Verdict != verdictSuppress {
			return fmt.Errorf("verdict for %q is %q, want %q or %q", v.Key, v.Verdict, verdictFile, verdictSuppress)
		}
		if v.Reason == "" {
			return fmt.Errorf("verdict for %q has an empty reason", v.Key)
		}
		if v.Verdict == verdictFile {
			files++
		}
	}
	date := when.UTC().Format(time.DateOnly)
	for _, v := range verdicts {
		item := queued[v.Key]
		rec := Record{Fingerprint: item.Fingerprint, Verdict: v.Verdict, Reason: v.Reason, Date: date}
		if v.Verdict == verdictFile {
			rec.Issue = v.Key
		}
		mem.Groups[v.Key] = rec
		mem.Resources[item.Resource] = when.UTC().Format(time.RFC3339)
	}
	return nil
}

func normalise(verdicts []Verdict) []Verdict {
	out := make([]Verdict, len(verdicts))
	for i, v := range verdicts {
		out[i] = Verdict{Key: strings.TrimSpace(v.Key), Verdict: strings.TrimSpace(v.Verdict), Reason: strings.TrimSpace(v.Reason)}
	}
	return out
}

const maxSummaryReason = 200

// summaryText collapses whitespace so an agent-written value cannot break the summary's markdown
// structure, and truncates it.
func summaryText(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if runes := []rune(s); len(runes) > maxSummaryReason {
		return string(runes[:maxSummaryReason]) + "..."
	}
	return s
}

func renderRecordSummary(verdicts []Verdict, when time.Time) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Kibana API gap verdicts recorded\n\n%d verdicts recorded at %s.\n\n", len(verdicts), when.UTC().Format(time.RFC3339))
	for _, v := range verdicts {
		fmt.Fprintf(&b, "- `%s`: %s (%s)\n", summaryText(v.Key), summaryText(v.Verdict), summaryText(v.Reason))
	}
	return b.String()
}

func cmdRecord(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportPath := fs.String("report", "", "report.json written by discover")
	verdictsPath := fs.String("verdicts", "", "verdicts.json written by the triage agent")
	memoryPath := fs.String("memory", "", "state.json to merge the verdicts into")
	at := fs.String("time", time.Now().UTC().Format(time.RFC3339), "verdict time (RFC 3339)")
	summaryPath := fs.String("summary", os.Getenv("GITHUB_STEP_SUMMARY"), "file to append a step summary to")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *reportPath == "" || *verdictsPath == "" || *memoryPath == "" {
		return errors.New("record requires --report, --verdicts and --memory")
	}
	when, err := time.Parse(time.RFC3339, *at)
	if err != nil {
		return fmt.Errorf("invalid --time %q: %w", *at, err)
	}
	var report Report
	if err := readJSON(*reportPath, &report, false); err != nil {
		return err
	}
	var verdicts []Verdict
	if err := readJSON(*verdictsPath, &verdicts, true); err != nil {
		return err
	}
	if len(verdicts) == 0 {
		return fmt.Errorf("%s holds no verdicts", *verdictsPath)
	}
	mem, err := loadMemory(*memoryPath)
	if err != nil {
		return err
	}
	if err := applyVerdicts(&report, mem, verdicts, when); err != nil {
		return err
	}
	if err := saveMemory(*memoryPath, mem); err != nil {
		return err
	}
	summary := renderRecordSummary(verdicts, when)
	fmt.Fprint(stdout, summary)
	return appendSummary(*summaryPath, summary)
}

func readJSON(path string, v any, strict bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return fmt.Errorf("parse %s: unexpected data after the JSON value", path)
	}
	return nil
}

func appendSummary(path, text string) error {
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, err = f.WriteString(text)
	return errors.Join(err, f.Close())
}
