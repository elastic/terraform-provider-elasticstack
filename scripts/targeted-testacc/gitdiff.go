// Package main implements a targeted acceptance test package selector.
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
