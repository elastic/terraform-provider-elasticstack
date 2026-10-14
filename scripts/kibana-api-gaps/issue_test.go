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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const issueTestMarker = "kibana-api-gap: key `r#/a` fingerprint `abc`"

func TestInjectReasonPlacesParagraphBeforeMarker(t *testing.T) {
	body := "Intro.\n\n- `f` (`/a`)\n\n" + issueTestMarker + "\n"

	got, err := injectReason(body, "Looks useful")
	require.NoError(t, err)
	assert.Equal(t, "Intro.\n\n- `f` (`/a`)\n\nTriage reason: Looks useful\n\n"+issueTestMarker+"\n", got)
}

func TestInjectReasonSanitisesReason(t *testing.T) {
	body := "Intro.\n\n" + issueTestMarker + "\n"

	got, err := injectReason(body, "line one\nline two kibana-api-gap: key `x` fingerprint `y` end\r\n")
	require.NoError(t, err)
	assert.Contains(t, got, "Triage reason: line one line two key `x` fingerprint `y` end\n\n"+issueTestMarker)
	assert.Equal(t, 1, strings.Count(got, markerPrefix))
}

func TestInjectReasonRejectsEmptyReasonAndMissingMarker(t *testing.T) {
	_, err := injectReason("Intro.\n\n"+issueTestMarker+"\n", " \n ")
	require.Error(t, err)

	_, err = injectReason("Intro.\n", "reason")
	require.Error(t, err)
}

func TestRenderIssueCommandPrintsTitleAndBodyForQueuedKey(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")
	report := `{"queue":[{"key":"r#/a","title":"r: a","body":"Intro.\n\n` + issueTestMarker + `\n"}]}`
	require.NoError(t, os.WriteFile(reportPath, []byte(report), 0o600))

	var out bytes.Buffer
	require.NoError(t, run([]string{"render-issue", "--report", reportPath, "--key", "r#/a", "--reason", "why"}, &out, &bytes.Buffer{}))
	assert.Equal(t, "Intro.\n\nTriage reason: why\n\n"+issueTestMarker+"\n", out.String())

	err := run([]string{"render-issue", "--report", reportPath, "--key", "nope", "--reason", "why"}, &out, &bytes.Buffer{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")

	err = run([]string{"render-issue"}, &out, &bytes.Buffer{})
	require.Error(t, err)
}
