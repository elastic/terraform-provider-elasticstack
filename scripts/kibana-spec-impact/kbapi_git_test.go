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
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gitHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

func TestGitShowPathOrMissingAndDiffKbapiAtRefs(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "test")
	genDir := filepath.Join(dir, "generated", "kbapi")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(genDir, "kibana.gen.go")
	if err := os.WriteFile(first, []byte("package kbapi\n\ntype OldOnly struct{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "with kbapi")
	oldHead := gitHead(t, dir)

	if err := os.Remove(first); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "remove kbapi")
	newHead := gitHead(t, dir)

	_, missingNew, err := gitShowPathOrMissing(dir, newHead)
	if err != nil || !missingNew {
		t.Fatalf("expect missing generated file at new head: missing=%v err=%v", missingNew, err)
	}
	content, missingOld, err := gitShowPathOrMissing(dir, oldHead)
	if err != nil || missingOld || !strings.Contains(content, "OldOnly") {
		t.Fatalf("old head content: missing=%v err=%v", missingOld, err)
	}

	changed, err := diffKbapiAtRefs(dir, oldHead, newHead)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(changed, "OldOnly") {
		t.Fatalf("expect removed type in diff, got %v", changed)
	}
}
