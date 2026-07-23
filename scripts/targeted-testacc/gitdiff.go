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
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type realGitDiffRunner struct{}

func (realGitDiffRunner) MergeBase() (string, error) {
	out, err := exec.Command("git", "merge-base", "origin/main", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (realGitDiffRunner) DiffNameOnly(base string) ([]string, error) {
	if base == "" {
		return nil, fmt.Errorf("diff base cannot be empty")
	}
	out, err := exec.Command("git", "diff", "--name-only", base+"..HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only %s..HEAD: %w", base, err)
	}
	return splitLines(string(out)), nil
}

// ResolveBaseline selects the diff baseline using the documented precedence:
//  1. explicit flag value
//  2. TARGETED_TESTACC_BASE environment variable
//  3. git merge-base origin/main HEAD
//  4. HEAD~1 fallback
func ResolveBaseline(flagBase string) string {
	if flagBase != "" {
		return flagBase
	}

	if envBase := os.Getenv("TARGETED_TESTACC_BASE"); envBase != "" {
		return envBase
	}

	if base, err := (realGitDiffRunner{}).MergeBase(); err == nil && base != "" {
		return base
	}

	return "HEAD~1"
}

// GitDiff returns the repository-relative changed file paths between the given
// baseline and HEAD.
func GitDiff(base string) ([]string, error) {
	return realGitDiffRunner{}.DiffNameOnly(base)
}

func splitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
