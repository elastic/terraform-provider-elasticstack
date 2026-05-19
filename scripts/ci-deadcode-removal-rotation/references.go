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
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

func runGoplsReferences(file string, line, col int) ([]string, error) {
	pos := fmt.Sprintf("%s:%d:%d", file, line, col)
	cmd := exec.Command("gopls", "references", pos)
	out, err := cmd.CombinedOutput()
	files, parseErr := parseGoplsReferencesOutput(bytes.NewReader(out))
	if parseErr != nil {
		return nil, parseErr
	}
	if err != nil && len(files) == 0 {
		return nil, fmt.Errorf("gopls references failed: %w", err)
	}
	return files, nil
}

func parseGoplsReferencesOutput(r io.Reader) ([]string, error) {
	var files []string
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		file := line[:idx]
		if file == "" {
			continue
		}
		if _, ok := seen[file]; !ok {
			seen[file] = struct{}{}
			files = append(files, file)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

// classifyReferences determines whether a candidate is safe to remove.
//
// A candidate is valid in two cases:
//   - 0 total references: pure dead code.
//   - All references come from exactly one *_test.go file that is in the SAME
//     package directory as the source file, and that file is not a black-box
//     acceptance test (acc_*_test.go).
//
// Anything else (non-test references, multiple test references, a single test
// reference in a different directory, or an acceptance test file) makes the
// candidate ineligible — the function is dead from the provider binary's
// point of view but is still needed.
func classifyReferences(srcFile string, refFiles []string) (eligible bool, testFile string) {
	srcFile = strings.TrimPrefix(srcFile, "./")

	testCount := 0
	var singleTestFile string
	for _, f := range refFiles {
		f = strings.TrimPrefix(f, "./")
		if strings.HasSuffix(f, "_test.go") {
			testCount++
			singleTestFile = f
		}
	}

	if testCount != len(refFiles) {
		return false, ""
	}
	if testCount == 0 {
		return true, ""
	}
	if testCount > 1 {
		return false, ""
	}

	if strings.HasPrefix(filepath.Base(singleTestFile), "acc_") {
		return false, ""
	}
	if filepath.Dir(singleTestFile) != filepath.Dir(srcFile) {
		return false, ""
	}
	return true, singleTestFile
}
