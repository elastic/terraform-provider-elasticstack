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
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

var breakingEndMarkerRE = regexp.MustCompile(`^\s*<!--\s*/breaking-changes\s*-->\s*$`)

// breakingHeadingAtLineStart mirrors /^###\s+Breaking changes/ used by the legacy heading detector for PR bodies.
var breakingHeadingAtLineStart = regexp.MustCompile(`^###\s+Breaking changes`)

// ExtractBreakingChanges returns the markdown content under ### Breaking changes (heading line excluded),
// or ("", false) when the heading is absent or only whitespace content remains after trimEnd.
// Semantics match the legacy JavaScript breaking extractor (fences + end marker + ##/### boundaries).
func ExtractBreakingChanges(changelogSection string) (string, bool) {
	if changelogSection == "" {
		return "", false
	}

	lineStrs := strings.Split(changelogSection, "\n")
	var inBreaking bool
	var fenceType byte // 0 = none, '`' = backtick fence, '~' = tilde fence
	var content []byte

	for _, ls := range lineStrs {
		line := []byte(ls)

		if breakingHeadingAtLineStart.Match(line) {
			inBreaking = true
			continue
		}
		if !inBreaking {
			continue
		}

		switch {
		case fenceType == 0 && bytes.HasPrefix(line, []byte("```")):
			fenceType = '`'
		case fenceType == 0 && bytes.HasPrefix(line, []byte("~~~")):
			fenceType = '~'
		case fenceType == '`' && bytes.HasPrefix(line, []byte("```")):
			fenceType = 0
		case fenceType == '~' && bytes.HasPrefix(line, []byte("~~~")):
			fenceType = 0
		}

		if fenceType == 0 && breakingEndMarkerRE.Match(line) {
			break
		}
		if fenceType == 0 && isHashHeadingLevel2Or3(line) {
			break
		}

		if len(content) > 0 {
			content = append(content, '\n')
		}
		content = append(content, line...)
	}

	if !inBreaking {
		return "", false
	}

	trimmed := trimRightMarkdownWhitespace(content)
	if len(trimmed) == 0 {
		return "", false
	}
	return string(trimmed), true
}

// isHashHeadingLevel2Or3 mirrors /^#{2,3}\s/.test(line) from the legacy breaking extractor boundaries.
func isHashHeadingLevel2Or3(line []byte) bool {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
		if i > 3 {
			return false
		}
	}
	if i < 2 || i > 3 {
		return false
	}
	if i >= len(line) {
		return false
	}
	c := line[i]
	return isMarkdownSpaceASCII(c)
}

func isMarkdownSpaceASCII(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

func trimRightMarkdownWhitespace(b []byte) []byte {
	return bytes.TrimRightFunc(b, unicode.IsSpace)
}

// BreakingHeadingPresent reports whether changelogSection contains a ### Breaking changes heading line.
func BreakingHeadingPresent(changelogSection string) bool {
	for ls := range strings.SplitSeq(changelogSection, "\n") {
		line := strings.TrimSuffix(ls, "\r")
		if breakingHeadingAtLineStart.MatchString(line) {
			return true
		}
	}
	return false
}
