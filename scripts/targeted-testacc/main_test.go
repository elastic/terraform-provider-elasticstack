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
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runToolFromModuleRoot executes `go run ./scripts/targeted-testacc/...` with
// the given arguments from the module root, as the Make target and workflow
// scripts do.
func runToolFromModuleRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root, err := moduleRoot(t)
	if err != nil {
		return "", err
	}
	cmd := exec.Command("go", append([]string{"run", "./scripts/targeted-testacc/..."}, args...)...)
	cmd.Dir = root
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	return string(out), err
}

func moduleRoot(t *testing.T) (string, error) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", os.ErrNotExist
		}
		wd = parent
	}
}

// decodePlan asserts stdout holds exactly one valid JSON shard plan.
func decodePlan(t *testing.T, stdout string) {
	t.Helper()

	var plan ShardPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &plan); err != nil {
		t.Fatalf("stdout is not a single JSON shard plan: %v\nstdout:\n%s", err, stdout)
	}
	if err := ValidateShardPlan(&plan); err != nil {
		t.Fatalf("tool emitted an invalid plan: %v\nplan: %#v", err, plan)
	}
	if len(plan.Rationale) == 0 {
		t.Fatalf("plan has no rationale: %#v", plan)
	}
}

func TestRunFromModuleRoot_EmitsPlan(t *testing.T) {
	stdout, err := runToolFromModuleRoot(t, "--total-shards=1")
	if err != nil {
		t.Fatalf("go run ./scripts/targeted-testacc/... from module root failed: %v\n%s", err, stdout)
	}
	decodePlan(t, stdout)
}

func TestRunFromModuleRoot_AllFlags(t *testing.T) {
	stdout, err := runToolFromModuleRoot(t, "--total-shards=2", "--base=origin/main", "--dry-run")
	if err != nil {
		t.Fatalf("tool failed with all flags: %v\n%s", err, stdout)
	}
	decodePlan(t, stdout)
}

func TestRunFromModuleRoot_RejectsInvalidShardCount(t *testing.T) {
	stdout, err := runToolFromModuleRoot(t, "--total-shards=0")
	if err == nil {
		t.Fatalf("tool exited 0 with --total-shards=0; stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, "has_packages") {
		t.Fatalf("tool emitted a shard plan despite an invalid shard count:\n%s", stdout)
	}
}

func TestRunFromModuleRoot_RejectsShardIndex(t *testing.T) {
	stdout, err := runToolFromModuleRoot(t, "--total-shards=2", "--shard-index=1")
	if err == nil {
		t.Fatalf("tool exited 0 with --shard-index; stdout:\n%s", stdout)
	}
	if strings.Contains(stdout, "has_packages") {
		t.Fatalf("tool emitted a shard plan despite --shard-index:\n%s", stdout)
	}
}
