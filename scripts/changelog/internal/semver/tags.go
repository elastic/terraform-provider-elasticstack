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

package semver

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	tagListGlob = `v[0-9]*.[0-9]*.[0-9]*`
	tagSortOpt  = "-version:refname"
)

// semverTagPattern matches changelog-release-context.js SEMVER_TAG_PATTERN.
var semverTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// Tag is a git tag pointing at a semver release (name includes leading "v").
type Tag string

// ListReleaseTags runs the same listing command as changelog-engine-factory.js TAG_LIST_CMD
// and filters to semver tags only (parseSemverTags).
func ListReleaseTags(execer Execer) ([]Tag, error) {
	out, err := execer.Run("git", "tag", "--list", tagListGlob, "--sort="+tagSortOpt)
	if err != nil {
		return nil, fmt.Errorf("git tag --list release tags: %w", err)
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return nil, nil
	}
	var tags []Tag
	for line := range strings.SplitSeq(raw, "\n") {
		t := strings.TrimSpace(line)
		if semverTagPattern.MatchString(t) {
			tags = append(tags, Tag(t))
		}
	}
	return tags, nil
}

// ParseSemverTagsFromRaw mirrors changelog-release-context.js parseSemverTags: split
// newline-separated input and keep only strict vX.Y.Z tags.
func ParseSemverTagsFromRaw(tagsRaw string) []Tag {
	var tags []Tag
	for line := range strings.SplitSeq(tagsRaw, "\n") {
		t := strings.TrimSpace(line)
		if semverTagPattern.MatchString(t) {
			tags = append(tags, Tag(t))
		}
	}
	return tags
}
