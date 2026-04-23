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
	"os/exec"
	"regexp"
	"strings"
)

var semverTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type releaseContext struct {
	Mode               string
	TargetVersion      string
	TargetBranch       string
	PreviousTag        string
	ExcludedTag        string
	ExcludedCurrentTag bool
	CompareRange       string
}

func listSemverTags() ([]string, error) {
	cmd := exec.Command("git", "tag", "--list", "v[0-9]*.[0-9]*.[0-9]*", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseSemverTags(string(output)), nil
}

func parseSemverTags(raw string) []string {
	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		tag := strings.TrimSpace(line)
		if semverTagPattern.MatchString(tag) {
			result = append(result, tag)
		}
	}
	return result
}

func resolveReleaseContext(mode, targetVersion string, tags []string) (releaseContext, error) {
	if mode == "" {
		return releaseContext{}, fmt.Errorf("mode is required")
	}
	if mode != changelogModeRelease && mode != "unreleased" {
		return releaseContext{}, fmt.Errorf("unsupported changelog mode: %s", mode)
	}
	if mode == changelogModeRelease && targetVersion == "" {
		return releaseContext{}, fmt.Errorf("release mode requires targetVersion")
	}

	ctx := releaseContext{
		Mode:          mode,
		TargetVersion: targetVersion,
		TargetBranch:  "generated-changelog",
		CompareRange:  "HEAD",
	}
	if mode == changelogModeRelease {
		ctx.TargetBranch = fmt.Sprintf("prep-release-%s", targetVersion)
		ctx.ExcludedTag = fmt.Sprintf("v%s", targetVersion)
	}

	candidates := make([]string, 0, len(tags))
	for _, tag := range tags {
		if ctx.ExcludedTag != "" && tag == ctx.ExcludedTag {
			ctx.ExcludedCurrentTag = true
			continue
		}
		candidates = append(candidates, tag)
	}
	if len(candidates) > 0 {
		ctx.PreviousTag = candidates[0]
		ctx.CompareRange = fmt.Sprintf("%s..HEAD", ctx.PreviousTag)
	}

	return ctx, nil
}
