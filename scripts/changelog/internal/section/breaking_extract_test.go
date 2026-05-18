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

package section

import (
	"strings"
	"testing"
)

func TestExtractBreaking_trimEndTreatsNbspLikeJS(t *testing.T) {
	section := "### Breaking changes\n\n\u00a0"
	_, ok := ExtractBreakingChanges(section)
	if ok {
		t.Fatal("NBSP-only body should trim to empty (parity with String.prototype.trimEnd)")
	}
}

func TestExtractBreaking_headingLevelFourDoesNotTerminate(t *testing.T) {
	section := strings.Join([]string{
		"### Breaking changes",
		"",
		"intro",
		"",
		"#### Notes",
		"",
		"must stay",
	}, "\n")
	out, ok := ExtractBreakingChanges(section)
	if !ok || !strings.Contains(out, "#### Notes") || !strings.Contains(out, "must stay") {
		t.Fatalf("got %q ok=%v", out, ok)
	}
}

func TestExtractBreaking_openBackticksCloseRequiresBackticksTwiceLikeJS(t *testing.T) {
	section := strings.Join([]string{
		"### Breaking changes",
		"",
		"```",
		"fenced-first-line",
		"~~~ is content until ``` closes backtick fence",
		"```",
	}, "\n")
	out, ok := ExtractBreakingChanges(section)
	if !ok {
		t.Fatal("expected content")
	}
	if !strings.Contains(out, "~~~") || !strings.Contains(out, "fenced-first-line") {
		t.Fatalf("got %q", out)
	}
}
