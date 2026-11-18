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

package memoryio

import (
	"os"
	"path/filepath"
	"testing"
)

type testPayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestAtomicWriteJSONAndReadJSONRoundTrip(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "memory.json")
	want := testPayload{Name: "foo", Count: 3}

	if err := AtomicWriteJSON(path, want, ".memoryio-*.json.tmp"); err != nil {
		t.Fatalf("AtomicWriteJSON: %v", err)
	}

	var got testPayload
	if err := ReadJSON(path, &got); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestAtomicWriteJSONCreatesParentDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "memory.json")

	if err := AtomicWriteJSON(path, testPayload{Name: "x"}, ".memoryio-*.json.tmp"); err != nil {
		t.Fatalf("AtomicWriteJSON: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}

func TestAtomicWriteJSONLeavesNoTempFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "memory.json")

	if err := AtomicWriteJSON(path, testPayload{Name: "x"}, ".memoryio-*.json.tmp"); err != nil {
		t.Fatalf("AtomicWriteJSON: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if e.Name() != "memory.json" {
			t.Errorf("unexpected leftover file: %s", e.Name())
		}
	}
}

func TestAtomicWriteJSONAppendsTrailingNewline(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "memory.json")
	if err := AtomicWriteJSON(path, testPayload{Name: "x"}, ".memoryio-*.json.tmp"); err != nil {
		t.Fatalf("AtomicWriteJSON: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Fatalf("expected trailing newline, got %q", data)
	}
}

func TestReadJSONMissingFile(t *testing.T) {
	t.Parallel()

	var got testPayload
	err := ReadJSON(filepath.Join(t.TempDir(), "missing.json"), &got)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadJSONInvalidJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var got testPayload
	if err := ReadJSON(path, &got); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
