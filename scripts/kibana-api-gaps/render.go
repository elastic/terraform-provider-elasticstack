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
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var versionPattern = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// introducedVersion extracts the Kibana version from an x-state value such as
// "Generally available in 8.15.0"; it is empty when the specification gives none.
func introducedVersion(xState string) string {
	return versionPattern.FindString(xState)
}

// renderTitle and renderBody produce the issue content from facts only. The `[kibana-api-gap] `
// title prefix is added by the create-issue safe output, and triage's reason is appended to the
// body by the agent (workflow task 4.1).
func renderTitle(g Group) string {
	return g.Resource + ": " + g.Field
}

// markerPrefix starts the machine-readable marker line at the end of every issue body.
// It is a visible footer, not an HTML comment with `key=`/`fp=` pairs: gh-aw's safe-output
// sanitizer strips HTML comments from issue bodies and its log redaction masks `key=` values.
const markerPrefix = "kibana-api-gap: key `"

// markerPattern reads the group key and fingerprint back from a marker line.
const markerPattern = "kibana-api-gap: key `([^\\s`]+)` fingerprint `([^\\s`]+)`"

// checkMarkerSafe verifies that the group key and fingerprint can be read back from the marker
// with markerPattern: non-empty, no whitespace and no backtick.
func checkMarkerSafe(g Group) error {
	for name, value := range map[string]string{"key": g.Key, "fingerprint": g.Fingerprint} {
		if value == "" || strings.ContainsFunc(value, unicode.IsSpace) || strings.Contains(value, "`") {
			return fmt.Errorf("group %s %q cannot be embedded in an issue marker", name, value)
		}
	}
	return nil
}

func marker(g Group) string {
	return fmt.Sprintf("%s%s` fingerprint `%s`", markerPrefix, g.Key, g.Fingerprint)
}

func renderBody(g Group) string {
	var b strings.Builder
	fmt.Fprintf(&b, "The Terraform resource `%s` never uses the following Kibana API request fields under `%s`.\n\n", g.Resource, g.Field)
	for _, m := range g.Members {
		fmt.Fprintf(&b, "- `%s` (`%s`)", m.Field, m.Pointer)
		if version := introducedVersion(m.XState); version != "" {
			fmt.Fprintf(&b, ", introduced in %s", version)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nThe locations are JSON pointers into the Kibana OpenAPI specification that `generated/kbapi` is generated from.\n\n")
	b.WriteString(marker(g))
	b.WriteString("\n")
	return b.String()
}
