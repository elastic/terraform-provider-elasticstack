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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const deadcodeTimeout = 15 * time.Minute

var deadcodeLinePattern = regexp.MustCompile(`^(.+):(\d+):(\d+): unreachable func: (.+)$`)

type deadcodeEntry struct {
	file        string
	line        int
	column      int
	symbol      string
	packagePath string
}

func (e deadcodeEntry) key() string {
	return e.packagePath + "." + e.symbol
}

// runDeadcode invokes go tool deadcode and parses its output.
func runDeadcode(testMode bool) ([]deadcodeEntry, error) {
	args := []string{"tool", "deadcode"}
	if testMode {
		args = append(args, "-test")
	}
	args = append(args, "./...")

	ctx, cancel := context.WithTimeout(context.Background(), deadcodeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	label := "deadcode"
	if testMode {
		label = "deadcode -test"
	}

	exitCode := -999
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	fmt.Fprintf(os.Stderr, "%s completed in %v (exit=%d stdout=%d bytes stderr=%d bytes)\n", label, elapsed, exitCode, stdout.Len(), stderr.Len())

	if stderr.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s stderr:\n%s\n", label, stderr.String())
	}

	entries, parseErr := parseDeadcodeOutput(bytes.NewReader(stdout.Bytes()))
	if parseErr != nil {
		return nil, parseErr
	}

	if len(entries) == 0 && stdout.Len() > 0 {
		fmt.Fprintf(os.Stderr, "%s stdout (first 2KB):\n%s\n", label, truncateBytes(stdout.Bytes(), 2048))
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("deadcode timed out after %v: %w", deadcodeTimeout, err)
		}
		if len(entries) == 0 {
			return nil, fmt.Errorf("deadcode failed: %w", err)
		}
	}
	return entries, nil
}

// excludedFilePrefixes lists relative file-path prefixes whose deadcode
// candidates should be ignored. These are test-only or dynamically-loaded
// packages that always appear dead from the provider main entrypoint.
var excludedFilePrefixes = []string{
	"analysis/acctestconfigdirlint/",
	"analysis/acctestconfigdirlintplugin/",
	"internal/acctest/",
	"internal/providerfwtest/",
	"internal/kibana/dashboard/panelkit/contracttest/",
}

func filterExcluded(entries []deadcodeEntry) []deadcodeEntry {
	var out []deadcodeEntry
	for _, e := range entries {
		if !isExcludedFile(e.file) {
			out = append(out, e)
		}
	}
	return out
}

func isExcludedFile(file string) bool {
	for _, p := range excludedFilePrefixes {
		if strings.HasPrefix(file, p) {
			return true
		}
	}
	return false
}

// hasExcludedReference returns true if any of the reference files lives in a
// known test/analysis package. A candidate whose only references are test-only
// code should be skipped because deadcode (run without -test) legitimately
// reports it dead, but removing it would break the test suite.
func hasExcludedReference(refFiles []string) bool {
	for _, f := range refFiles {
		if isExcludedFile(f) {
			return true
		}
	}
	return false
}

func truncateBytes(b []byte, max int) string {
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + " [...truncated]"
}

func parseDeadcodeOutput(r io.Reader) ([]deadcodeEntry, error) {
	var entries []deadcodeEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		m := deadcodeLinePattern.FindStringSubmatch(line)
		if len(m) != 5 {
			continue
		}
		lineNum, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("invalid line number %q: %w", m[2], err)
		}
		colNum, err := strconv.Atoi(m[3])
		if err != nil {
			return nil, fmt.Errorf("invalid column number %q: %w", m[3], err)
		}
		entries = append(entries, deadcodeEntry{
			file:   m[1],
			line:   lineNum,
			column: colNum,
			symbol: m[4],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func derivePackagePath(filePath, modulePath string) string {
	dir := filepath.Dir(filePath)
	if dir == "." {
		return modulePath
	}
	return modulePath + "/" + filepath.ToSlash(dir)
}

func intersectCandidates(a, b []deadcodeEntry) []deadcodeEntry {
	set := make(map[string]struct{}, len(b))
	for _, e := range b {
		set[e.key()] = struct{}{}
	}
	var out []deadcodeEntry
	for _, e := range a {
		if _, ok := set[e.key()]; ok {
			out = append(out, e)
		}
	}
	return out
}

func selectEligible(candidates []deadcodeEntry, mem *Memory, now time.Time) []deadcodeEntry {
	var eligible []deadcodeEntry
	for _, c := range candidates {
		if !isInCooldown(mem, c.key(), now) {
			eligible = append(eligible, c)
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].key() < eligible[j].key()
	})
	return eligible
}

// selectOne is kept for backward compatibility.
func selectOne(candidates []deadcodeEntry, mem *Memory, now time.Time) *deadcodeEntry {
	eligible := selectEligible(candidates, mem, now)
	if len(eligible) == 0 {
		return nil
	}
	return &eligible[0]
}

type Candidate struct {
	Symbol                       string   `json:"symbol"`
	SymbolName                   string   `json:"symbol_name"`
	Package                      string   `json:"package"`
	File                         string   `json:"file"`
	Line                         int      `json:"line"`
	Column                       int      `json:"column"`
	CompanionTestCleanupEligible bool     `json:"companion_test_cleanup_eligible"`
	CompanionTestFile            string   `json:"companion_test_file"`
	ReferenceFiles               []string `json:"reference_files"`
	ImpactedPackages             []string `json:"impacted_packages"`
	Found                        bool     `json:"found"`
}

func impactedPackages(entry deadcodeEntry, testFile string) []string {
	pkgs := map[string]struct{}{}
	dir := filepath.Dir(entry.file)
	if dir == "." {
		pkgs["."] = struct{}{}
	} else {
		pkgs["./"+filepath.ToSlash(dir)] = struct{}{}
	}
	if testFile != "" {
		tdir := filepath.Dir(testFile)
		if tdir != dir {
			if tdir == "." {
				pkgs["."] = struct{}{}
			} else {
				pkgs["./"+filepath.ToSlash(tdir)] = struct{}{}
			}
		}
	}
	var out []string
	for p := range pkgs {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
