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
	"strings"
	"testing"
)

// fakeGoShim exercises the Make target without toolchain builds or live tests.
func fakeGoShim(t *testing.T, plan string) (logPath string) {
	t.Helper()

	dir := t.TempDir()
	logPath = filepath.Join(dir, "go-invocations.log")

	shim := "#!/usr/bin/env bash\n" +
		"set -u\n" +
		"if [ \"${1:-}\" = run ]; then\n" +
		"printf '%s\\n' \"run $*\" >>\"" + logPath + "\"\n" +
		"printf '%s\\n' \"$GO_SHIM_PLAN\"\n" +
		"exit 0\n" +
		"fi\n" +
		"printf '%s\\n' \"TF_ACC=${TF_ACC:-} $*\" >>\"" + logPath + "\"\n" +
		"exit 0\n"

	bin := filepath.Join(dir, "go")
	if err := os.WriteFile(bin, []byte(shim), 0o755); err != nil {
		t.Fatalf("write go shim: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("GO_SHIM_PLAN", plan)
	return logPath
}

func runMakeTargetedTestacc(t *testing.T, dir string, extraEnv ...string) string {
	t.Helper()

	cmd := exec.Command("make", "targeted-testacc")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("make targeted-testacc failed: %v\nstdout:\n%s", err, out)
	}
	return string(out)
}

func TestMakeTarget_NoPackagesSelectedExitsCleanly(t *testing.T) {
	root := repoRoot(t)
	plan := `{"has_packages":false,"selected_packages":[],"shards":[[]],"rationale":["no resolvable diff"]}`
	shimLog := fakeGoShim(t, plan)

	stdout := runMakeTargetedTestacc(t, root, "ACCTEST_SHARD_INDEX=0", "ACCTEST_TOTAL_SHARDS=1")
	if !strings.Contains(stdout, "No acceptance test packages selected") {
		t.Errorf("make printed no empty-selection notice; stdout:\n%s", stdout)
	}

	logBytes, err := os.ReadFile(shimLog)
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.SplitSeq(strings.TrimSpace(string(logBytes)), "\n") {
		if strings.Contains(line, "gotestsum") {
			t.Errorf("gotestsum invoked for an empty shard: %s", line)
		}
	}
}

func TestMakeTarget_InvalidShardIndexFails(t *testing.T) {
	plan := `{"has_packages":true,"selected_packages":["pkg/a","pkg/b"],"shards":[["pkg/a"],["pkg/b"]],"rationale":["diff"]}`

	for name, env := range map[string]string{
		"non-integer": "ACCTEST_SHARD_INDEX=not-a-number",
		"negative":    "ACCTEST_SHARD_INDEX=-1",
		"blank":       "ACCTEST_SHARD_INDEX=",
		"whitespace":  "ACCTEST_SHARD_INDEX= ",
		"float":       "ACCTEST_SHARD_INDEX=1.5",
	} {
		t.Run(name, func(t *testing.T) {
			root := repoRoot(t)
			fakeGoShim(t, plan)

			cmd := exec.Command("make", "targeted-testacc")
			cmd.Dir = root
			cmd.Env = append(os.Environ(), env, "ACCTEST_TOTAL_SHARDS=2")
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("make targeted-testacc succeeded with %s\nstdout:\n%s", env, out)
			}
			if !strings.Contains(string(out), "ACCTEST_SHARD_INDEX") {
				t.Errorf("make failure did not name ACCTEST_SHARD_INDEX; output:\n%s", out)
			}
		})
	}
}

func TestMakeTarget_PackagesSelectedRunsGotestsum(t *testing.T) {
	root := repoRoot(t)
	plan := `{"has_packages":true,"selected_packages":["pkg/a","pkg/b"],"shards":[["pkg/a"],["pkg/b"]],"rationale":["diff"]}`
	shimLog := fakeGoShim(t, plan)

	stdout := runMakeTargetedTestacc(t, root, "ACCTEST_SHARD_INDEX=1", "ACCTEST_TOTAL_SHARDS=2")
	if strings.Contains(stdout, "No acceptance test packages selected") {
		t.Errorf("make skipped a populated shard; stdout:\n%s", stdout)
	}

	logBytes, err := os.ReadFile(shimLog)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logBytes), "TF_ACC=1 tool gotestsum") {
		t.Errorf("gotestsum not invoked with TF_ACC=1 for a populated shard; log:\n%s", logBytes)
	}
	if !strings.Contains(string(logBytes), "--packages=pkg/b") {
		t.Errorf("gotestsum did not receive shard 1 packages; log:\n%s", logBytes)
	}
}

func TestMakeTarget_TargetedPkgsBypassesSelectionTool(t *testing.T) {
	root := repoRoot(t)
	shimLog := fakeGoShim(t, `{}`)

	pkgs := "github.com/example/mod/internal/a github.com/example/mod/internal/b"
	runMakeTargetedTestacc(t, root, "TARGETED_PKGS="+pkgs)

	logBytes, err := os.ReadFile(shimLog)
	if err != nil {
		t.Fatal(err)
	}
	log := string(logBytes)

	for line := range strings.SplitSeq(strings.TrimSpace(log), "\n") {
		if strings.HasPrefix(line, "run ") {
			t.Errorf("selection tool invoked despite TARGETED_PKGS: %s", line)
		}
	}
	if !strings.Contains(log, "--packages="+pkgs) {
		t.Errorf("gotestsum did not receive TARGETED_PKGS verbatim; log:\n%s", log)
	}
}
