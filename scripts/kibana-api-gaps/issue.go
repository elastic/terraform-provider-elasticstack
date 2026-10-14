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
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// injectReason inserts a single `Triage reason:` paragraph immediately before the hidden marker.
// The reason is flattened to one line and stripped of the marker prefix so it cannot forge a marker.
func injectReason(body, reason string) (string, error) {
	reason = strings.ReplaceAll(reason, "kibana-api-gap:", "")
	reason = strings.Join(strings.Fields(reason), " ")
	if reason == "" {
		return "", errors.New("reason is empty")
	}
	at := strings.LastIndex(body, markerPrefix)
	if at < 0 {
		return "", errors.New("issue body has no kibana-api-gap marker")
	}
	return body[:at] + "Triage reason: " + reason + "\n\n" + body[at:], nil
}

// cmdRenderIssue prints the final issue body for a queued key, so the agent never edits the
// pre-rendered body itself.
func cmdRenderIssue(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("render-issue", flag.ContinueOnError)
	fs.SetOutput(stderr)
	reportPath := fs.String("report", "", "report.json written by discover")
	key := fs.String("key", "", "queue item key")
	reason := fs.String("reason", "", "triage reason")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *reportPath == "" || *key == "" {
		return errors.New("render-issue requires --report, --key and --reason")
	}
	var report Report
	if err := readJSON(*reportPath, &report, false); err != nil {
		return err
	}
	for _, item := range report.Queue {
		if item.Key == *key {
			body, err := injectReason(item.Body, *reason)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(stdout, body)
			return err
		}
	}
	return fmt.Errorf("key %q is not in the triage queue", *key)
}
