// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
//
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

import "fmt"

// PreviousTagResult mirrors selectPreviousTag in changelog-release-context.js.
type PreviousTagResult struct {
	PreviousTag        string
	ExcludedTag        string
	ExcludedCurrentTag bool
}

const (
	headCompareRef              = "HEAD"
	rangeSep                    = ".."
	changelogReleaseModeLiteral = "release"
)

// SelectPreviousTag implements changelog-release-context.js selectPreviousTag.
func SelectPreviousTag(tags []Tag, mode, targetVersion string) PreviousTagResult {
	excludedTag := ""
	if mode == changelogReleaseModeLiteral && targetVersion != "" {
		excludedTag = fmt.Sprintf("v%s", targetVersion)
	}

	candidates := tags
	var filtered []Tag
	if excludedTag != "" {
		for _, t := range tags {
			if string(t) != excludedTag {
				filtered = append(filtered, t)
			}
		}
		candidates = filtered
	}

	previous := ""
	if len(candidates) > 0 {
		previous = string(candidates[0])
	}

	excludedCurrent := excludedTag != "" && len(candidates) < len(tags)

	return PreviousTagResult{
		PreviousTag:        previous,
		ExcludedTag:        excludedTag,
		ExcludedCurrentTag: excludedCurrent,
	}
}

// BuildCompareRange implements changelog-release-context.js buildCompareRange ("base..HEAD").
func BuildCompareRange(previousTag string) string {
	if previousTag == "" {
		return headCompareRef
	}
	return previousTag + rangeSep + headCompareRef
}
