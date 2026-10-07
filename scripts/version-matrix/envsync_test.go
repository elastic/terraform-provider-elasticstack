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
	"path/filepath"
	"testing"
)

func TestLatestGA(t *testing.T) {
	got, err := LatestGA([]string{"8.19.23", "9.4.8", "9.10.1", "9.5.5", "9.11.0-SNAPSHOT"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "9.10.1" {
		t.Fatalf("got %q, want 9.10.1", got)
	}
	if _, err := LatestGA([]string{"9.6.0-SNAPSHOT"}); err == nil {
		t.Fatal("expected error when only snapshots are present")
	}
}

func TestSetStackVersion(t *testing.T) {
	in := "STACK_VERSION=9.4.0\nELASTICSEARCH_PASSWORD=password\n"
	out, changed, err := SetStackVersion(in, "9.5.5")
	if err != nil || !changed {
		t.Fatalf("changed=%v err=%v", changed, err)
	}
	if want := "STACK_VERSION=9.5.5\nELASTICSEARCH_PASSWORD=password\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
	_, changed, err = SetStackVersion(out, "9.5.5")
	if err != nil || changed {
		t.Fatalf("expected idempotent no-op, changed=%v err=%v", changed, err)
	}
	if _, _, err := SetStackVersion("FOO=bar\n", "9.5.5"); err == nil {
		t.Fatal("expected error when STACK_VERSION line is missing")
	}
}

func TestSyncEnvFile(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "matrix.json")
	envFile := filepath.Join(dir, ".env.template")
	if err := os.WriteFile(artifact, []byte(`["9.4.8","9.5.5","9.6.0-SNAPSHOT"]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(envFile, []byte("STACK_VERSION=9.4.0\nOTHER=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v, changed, err := SyncEnvFile(artifact, envFile)
	if err != nil || !changed || v != "9.5.5" {
		t.Fatalf("v=%q changed=%v err=%v", v, changed, err)
	}
	data, _ := os.ReadFile(envFile)
	if string(data) != "STACK_VERSION=9.5.5\nOTHER=1\n" {
		t.Fatalf("unexpected content %q", data)
	}
	if _, changed, err = SyncEnvFile(artifact, envFile); err != nil || changed {
		t.Fatalf("second run changed=%v err=%v", changed, err)
	}
}
