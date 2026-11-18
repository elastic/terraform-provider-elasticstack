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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMemoryTreatsAMissingFileAsEmpty(t *testing.T) {
	mem, err := loadMemory(filepath.Join(t.TempDir(), "absent", "state.json"))
	require.NoError(t, err)
	assert.Empty(t, mem.Groups)
	assert.Empty(t, mem.Resources)
}

func TestLoadMemoryRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))
	_, err := loadMemory(path)
	require.Error(t, err)
}

func TestSaveMemoryCreatesDirectoriesAndRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory", "branch", "state.json")
	mem := &Memory{
		Groups:    map[string]Record{"r#/p": {Fingerprint: "f", Verdict: verdictSuppress, Reason: "server-managed", Date: "2026-10-13"}},
		Resources: map[string]string{"r": "2026-10-13"},
	}

	require.NoError(t, saveMemory(path, mem))

	got, err := loadMemory(path)
	require.NoError(t, err)
	assert.Equal(t, mem, got)
}

func TestSaveMemoryIsByteStable(t *testing.T) {
	dir := t.TempDir()
	mem := &Memory{
		Groups:    map[string]Record{"b#/p": {Fingerprint: "f"}, "a#/p": {Fingerprint: "f"}},
		Resources: map[string]string{"b": "2026-01-01", "a": "2026-01-01"},
	}
	require.NoError(t, saveMemory(filepath.Join(dir, "1.json"), mem))
	require.NoError(t, saveMemory(filepath.Join(dir, "2.json"), mem))
	first, err := os.ReadFile(filepath.Join(dir, "1.json"))
	require.NoError(t, err)
	second, err := os.ReadFile(filepath.Join(dir, "2.json"))
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestMemoryCoversAGroupOnlyWhileTheFingerprintMatches(t *testing.T) {
	mem := &Memory{Groups: map[string]Record{"k": {Fingerprint: "a"}}}
	assert.True(t, mem.covers(Group{Key: "k", Fingerprint: "a"}))
	assert.False(t, mem.covers(Group{Key: "k", Fingerprint: "b"}))
	assert.False(t, mem.covers(Group{Key: "other", Fingerprint: "a"}))
	assert.False(t, (&Memory{}).covers(Group{Key: "k"}))
}

func TestLoadMemoryReplacesNullMapsWithEmptyOnes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"groups":null,"resources":null}`), 0o600))

	mem, err := loadMemory(path)

	require.NoError(t, err)
	assert.NotNil(t, mem.Groups)
	assert.NotNil(t, mem.Resources)
	mem.Groups["k"] = Record{}
}

func TestSaveMemoryFailsWhenTheDirectoryCannotBeCreated(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(blocker, nil, 0o600))
	require.Error(t, saveMemory(filepath.Join(blocker, "state.json"), newMemory()))
}

func TestSaveMemoryLeavesNoTemporaryFilesAndKeepsTheOldFileOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	require.NoError(t, saveMemory(path, newMemory()))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "state.json", entries[0].Name())
}
