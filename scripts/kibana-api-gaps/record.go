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

// applyVerdicts validates verdicts against the report's queue and merges them into mem. mem is
// only modified when every verdict is valid.
func applyVerdicts(report *Report, mem *Memory, verdicts []Verdict, date string) error {
	queued := make(map[string]QueueItem, len(report.Queue))
	for _, item := range report.Queue {
		queued[item.Key] = item
	}
	seen := make(map[string]bool, len(verdicts))
	files := 0
	for _, v := range verdicts {
		if _, ok := queued[v.Key]; !ok {
			return fmt.Errorf("verdict for unknown key %q: it is not in the triage queue", v.Key)
		}
		if seen[v.Key] {
			return fmt.Errorf("duplicate verdict for key %q", v.Key)
		}
		seen[v.Key] = true
		if v.Verdict != verdictFile && v.Verdict != verdictSuppress {
			return fmt.Errorf("verdict for %q is %q, want %q or %q", v.Key, v.Verdict, verdictFile, verdictSuppress)
		}
		if strings.TrimSpace(v.Reason) == "" {
			return fmt.Errorf("verdict for %q has an empty reason", v.Key)
		}
		if v.Verdict == verdictFile {
			files++
		}
	}
	if files > report.Slots {
		return fmt.Errorf("%d file verdicts exceed the %d available slots", files, report.Slots)
	}
	for _, v := range verdicts {
		item := queued[v.Key]
		rec := Record{Fingerprint: item.Fingerprint, Verdict: v.Verdict, Reason: v.Reason, Date: date}
		if v.Verdict == verdictFile {
			rec.Issue = v.Key
		}
		mem.Groups[v.Key] = rec
		mem.Resources[item.Resource] = date
	}
	return nil
}

func renderRecordSummary(verdicts []Verdict, date string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Kibana API gap verdicts recorded\n\n%d verdicts recorded on %s.\n\n", len(verdicts), date)
	for _, v := range verdicts {
		fmt.Fprintf(&b, "- `%s`: %s (%s)\n", v.Key, v.Verdict, v.Reason)
	}
	return b.String()
}

func cmdRecord(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportPath := fs.String("report", "", "report.json written by discover")
	verdictsPath := fs.String("verdicts", "", "verdicts.json written by the triage agent")
	memoryPath := fs.String("memory", "", "state.json to merge the verdicts into")
	date := fs.String("date", time.Now().UTC().Format(time.DateOnly), "verdict date (YYYY-MM-DD)")
	summaryPath := fs.String("summary", os.Getenv("GITHUB_STEP_SUMMARY"), "file to append a step summary to")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *reportPath == "" || *verdictsPath == "" || *memoryPath == "" {
		return errors.New("record requires --report, --verdicts and --memory")
	}
	var report Report
	if err := readJSON(*reportPath, &report, false); err != nil {
		return err
	}
	var verdicts []Verdict
	if err := readJSON(*verdictsPath, &verdicts, true); err != nil {
		return err
	}
	mem, err := loadMemory(*memoryPath)
	if err != nil {
		return err
	}
	if err := applyVerdicts(&report, mem, verdicts, *date); err != nil {
		return err
	}
	if err := saveMemory(*memoryPath, mem); err != nil {
		return err
	}
	summary := renderRecordSummary(verdicts, *date)
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
