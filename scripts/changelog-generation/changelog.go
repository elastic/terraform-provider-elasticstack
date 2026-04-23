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
	"strings"
	"time"
)

const changelogModeRelease = "release"

func buildSectionContent(mode, targetVersion string, generatedAt time.Time, sectionBody string) (string, string) {
	sectionHeader := "## [Unreleased]"
	if mode == changelogModeRelease && targetVersion != "" {
		sectionHeader = fmt.Sprintf("## [%s] - %s", targetVersion, generatedAt.Format("2006-01-02"))
	}
	newSectionContent := sectionHeader
	if sectionBody != "" {
		newSectionContent = sectionHeader + "\n\n" + sectionBody
	}
	return sectionHeader, newSectionContent
}

func findSectionEnd(lines []string, startIndex int) int {
	for i := startIndex + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			return i
		}
	}
	return len(lines)
}

func rewriteChangelogSection(content, newSectionContent, mode, targetVersion string) string {
	lines := strings.Split(content, "\n")
	targetStart := -1
	if mode == "unreleased" {
		for i, line := range lines {
			if strings.HasPrefix(line, "## [Unreleased]") {
				targetStart = i
				break
			}
		}
	} else {
		prefix := fmt.Sprintf("## [%s]", targetVersion)
		for i, line := range lines {
			if strings.HasPrefix(line, prefix) {
				targetStart = i
				break
			}
		}
	}

	if targetStart == -1 {
		if mode == changelogModeRelease {
			unreleasedStart := -1
			for i, line := range lines {
				if strings.HasPrefix(line, "## [Unreleased]") {
					unreleasedStart = i
					break
				}
			}
			if unreleasedStart != -1 {
				insertAfter := findSectionEnd(lines, unreleasedStart)
				before := append([]string{}, lines[:insertAfter]...)
				after := append([]string{}, lines[insertAfter:]...)
				return strings.Join(append(append(before, "", newSectionContent), after...), "\n")
			}
		}
		if content == "" {
			return newSectionContent
		}
		return newSectionContent + "\n\n" + content
	}

	sectionEnd := findSectionEnd(lines, targetStart)
	before := append([]string{}, lines[:targetStart]...)
	after := append([]string{}, lines[sectionEnd:]...)
	for len(before) > 0 && before[len(before)-1] == "" {
		before = before[:len(before)-1]
	}
	parts := append([]string{}, before...)
	if len(parts) > 0 {
		parts = append(parts, "")
	}
	parts = append(parts, newSectionContent)
	afterStart := 0
	for afterStart < len(after) && after[afterStart] == "" {
		afterStart++
	}
	if afterStart < len(after) {
		parts = append(parts, "")
		parts = append(parts, after[afterStart:]...)
	}
	return strings.Join(parts, "\n")
}
