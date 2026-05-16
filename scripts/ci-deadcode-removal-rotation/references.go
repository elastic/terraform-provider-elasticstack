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
	if err != nil {
		if len(out) == 0 {
			return nil, fmt.Errorf("gopls references failed: %w", err)
		}
	}
	return parseGoplsReferencesOutput(bytes.NewReader(out))
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

func relativePath(base, target string) (string, error) {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	return rel, nil
}

func classifyReferences(refFiles []string) (eligible bool, testFile string) {
	var testFiles []string
	for _, f := range refFiles {
		if strings.HasSuffix(f, "_test.go") {
			testFiles = append(testFiles, f)
		}
	}
	if len(testFiles) == 1 {
		if strings.HasPrefix(filepath.Base(testFiles[0]), "acc_") {
			return false, ""
		}
		return true, testFiles[0]
	}
	return false, ""
}
