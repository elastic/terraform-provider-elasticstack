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

func (realGitDiffRunner) DiffNameOnly(base string, twoDot bool) ([]string, error) {
	if base == "" {
		return nil, fmt.Errorf("diff base cannot be empty")
	}
	// When the baseline was supplied explicitly (e.g. the fetched PR base
	// commit in CI), use the two-dot form: the PR merge ref already contains
	// the base, so <base>..HEAD is the correct PR diff and resolves even in a
	// shallow checkout where the merge base of <base> and HEAD is unreachable.
	// For auto-detected (moving) baselines such as origin/main, the three-dot
	// (merge-base) form avoids over-selecting changes that have already been
	// merged; those paths are locally resolvable by construction.
	sep := ".."
	if !twoDot {
		sep = "..."
	}
	out, err := exec.Command("git", "diff", "--name-only", base+sep+"HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("git diff --name-only %s%sHEAD: %w", base, sep, err)
	}
	return splitLines(string(out)), nil
}

// Baseline is the resolved diff baseline plus whether it was supplied
// explicitly (flag or environment variable).
type Baseline struct {
	Ref string
	// Explicit is true when the baseline came from --base or
	// TARGETED_TESTACC_BASE. Explicit baselines are expected to be commits
	// that are already contained in HEAD (e.g. the PR base), so the diff uses
	// the two-dot form which resolves even in shallow checkouts.
	Explicit bool
}

// ResolveBaseline selects the diff baseline using the documented precedence:
//  1. explicit flag value
//  2. TARGETED_TESTACC_BASE environment variable
//  3. git merge-base origin/main HEAD
//  4. HEAD~1 fallback
func ResolveBaseline(flagBase string) Baseline {
	if flagBase != "" {
		return Baseline{Ref: flagBase, Explicit: true}
	}

	if envBase := os.Getenv("TARGETED_TESTACC_BASE"); envBase != "" {
		return Baseline{Ref: envBase, Explicit: true}
	}

	if base, err := (realGitDiffRunner{}).MergeBase(); err == nil && base != "" {
		return Baseline{Ref: base}
	}

	return Baseline{Ref: "HEAD~1"}
}

// GitDiff returns the repository-relative changed file paths between the given
// baseline and HEAD.
func GitDiff(base Baseline) ([]string, error) {
	return realGitDiffRunner{}.DiffNameOnly(base.Ref, base.Explicit)
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
