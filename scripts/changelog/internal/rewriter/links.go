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

package rewriter

import (
	"regexp"
)

var unreleasedLinkLine = regexp.MustCompile(`(?m)^\[Unreleased\]:[ \t]*(https?://.+/compare/).*$`)

var unreleasedWholeLine = regexp.MustCompile(`(?m)^\[Unreleased\]:.*$`)

// LinkEntry is one release bump for the changelog compare link table ([Unreleased]: / [x.y.z]: lines).
type LinkEntry struct {
	TargetVersion string
	PreviousTag   string // with leading 'v', e.g. v0.14.5
}

// UpdateLinks applies rewriteLinkTable from changelog-rewriter.js for each entry (sequentially).
//
// Behaviour details (mirrors JS exactly for one entry — the common case): only the [Unreleased]
// line anchors mutation; inserting a missing [targetVersion]: line nests it adjacent to Unreleased,
// preserving ordering of unrelated link lines afterward.
func UpdateLinks(content []byte, entries []LinkEntry) ([]byte, error) {
	out := string(content)
	for _, e := range entries {
		out = rewriteLinkTable(out, e.TargetVersion, e.PreviousTag)
	}
	return []byte(out), nil
}

func rewriteLinkTable(content, targetVersion, previousTag string) string {
	if targetVersion == "" || previousTag == "" {
		return content
	}

	unreleasedMatch := unreleasedLinkLine.FindStringSubmatch(content)
	if unreleasedMatch == nil {
		return content
	}

	baseCompareURL := unreleasedMatch[1]
	unreleasedLine := "[Unreleased]: " + baseCompareURL + "v" + targetVersion + "...HEAD"
	releaseLine := "[" + targetVersion + "]: " + baseCompareURL + previousTag + "...v" + targetVersion

	updated := replaceFirstMatch(content, unreleasedLinkLine, unreleasedLine)

	if versionLinkExists(updated, targetVersion) {
		return updated
	}

	return replaceFirstMatch(updated, unreleasedWholeLine, unreleasedLine+"\n"+releaseLine)
}

func replaceFirstMatch(s string, re *regexp.Regexp, replacement string) string {
	loc := re.FindStringIndex(s)
	if loc == nil {
		return s
	}
	return s[:loc[0]] + replacement + s[loc[1]:]
}

func versionLinkExists(updated, targetVersion string) bool {
	re := regexp.MustCompile(`(?m)^\[` + regexp.QuoteMeta(targetVersion) + `\]:`)
	return re.MatchString(updated)
}
