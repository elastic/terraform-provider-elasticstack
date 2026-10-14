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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderIssueTitleNamesResourceAndTopLevelField(t *testing.T) {
	g := gapGroup("elasticstack_kibana_slo", "settings", 1, "f")
	assert.Equal(t, "elasticstack_kibana_slo: settings", renderTitle(g))
}

func TestRenderIssueBodyListsEveryFieldWithLocationAndVersion(t *testing.T) {
	g := gapGroup("res", "top", 0, "fp")
	g.Members = []Member{
		{Field: "top.a", Pointer: "/components/schemas/W/properties/top/properties/a", XState: "Generally available in 8.15.0"},
		{Field: "top.b", Pointer: "/components/schemas/W/properties/top/properties/b"},
	}

	body := renderBody(g)

	assert.Contains(t, body, "`res`")
	assert.Contains(t, body, "`top.a`")
	assert.Contains(t, body, "`/components/schemas/W/properties/top/properties/a`")
	assert.Contains(t, body, "introduced in 8.15.0")
	assert.Contains(t, body, "`top.b`")
	assert.Equal(t, 1, countOccurrences(body, "introduced in"), "no version for a field without one")
	assert.Contains(t, body, "kibana-api-gap: key `"+g.Key+"` fingerprint `fp`")
	// gh-aw's safe-output sanitizer strips HTML comments and the log redactor masks `key=` values,
	// so the marker must be neither.
	assert.NotContains(t, body, "<!--")
	assert.NotContains(t, body, "key=")
	assert.NotContains(t, body, "fp=")
	// The filtered specification is not available to the triage agent, so the body must not send
	// it looking for the file.
	assert.NotContains(t, body, "oas-filtered")
}

func TestIntroducedVersionExtractsTheVersionFromXState(t *testing.T) {
	for state, want := range map[string]string{
		"Generally available in 8.15.0": "8.15.0",
		"Technical Preview in 9.1":      "9.1",
		"Beta":                          "",
		"":                              "",
	} {
		assert.Equal(t, want, introducedVersion(state), state)
	}
}

func countOccurrences(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}

func TestCheckMarkerSafeRejectsKeysThatCouldBreakTheMarker(t *testing.T) {
	assert.NoError(t, checkMarkerSafe(gapGroup("r", "a", 1, "fp")))
	assert.NoError(t, checkMarkerSafe(Group{Key: "r#/paths/~1api~1x~1{id}/get/parameters/format", Fingerprint: "fp"}))
	for _, key := range []string{"r#/a b", "r#/a`b", "r#/a\n", ""} {
		assert.Error(t, checkMarkerSafe(Group{Key: key, Fingerprint: "fp"}), key)
	}
}

func TestMarkerRoundTripsThroughTheDocumentedPattern(t *testing.T) {
	g := gapGroup("r", "a", 1, "abc123")
	match := regexp.MustCompile(markerPattern).FindStringSubmatch(renderBody(g))
	require.Len(t, match, 3)
	assert.Equal(t, g.Key, match[1])
	assert.Equal(t, "abc123", match[2])
}
