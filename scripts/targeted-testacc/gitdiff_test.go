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
	"testing"
)

func TestRealGitDiffRunner_DiffNameOnly_ListsBothPathsOfRename(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	if err := os.MkdirAll(filepath.Join(dir, "pkga"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkga", "a.go"), []byte("package pkga\n// content long enough to be detected as a rename\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "base")
	run("tag", "base")
	if err := os.MkdirAll(filepath.Join(dir, "pkgb"), 0o755); err != nil {
		t.Fatal(err)
	}
	run("mv", "pkga/a.go", "pkgb/a.go")
	run("commit", "-q", "-m", "rename")

	t.Chdir(dir)
	got, err := realGitDiffRunner{}.DiffNameOnly("base", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pkga/a.go", "pkgb/a.go"} {
		if !slices.Contains(got, want) {
			t.Errorf("DiffNameOnly = %v, missing %s", got, want)
		}
	}
}
