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
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunLoadMatrixWritesGitHubOutputsFromArtifact(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"8.1.3", "8.10.4"}))

	outPath := filepath.Join(dir, "github-output")
	t.Setenv("GITHUB_OUTPUT", outPath)

	err := run([]string{"load-matrix", "-artifact", artifact}, io.Discard, io.Discard)
	require.NoError(t, err)

	outputs := parseGitHubOutput(t, outPath)
	var versions []string
	require.NoError(t, json.Unmarshal([]byte(outputs["versions"]), &versions))
	assert.Equal(t, []string{"8.1.3", "8.10.4"}, versions)

	var flags map[string]MatrixEntryFlags
	require.NoError(t, json.Unmarshal([]byte(outputs["flags"]), &flags))
	assert.Equal(t, FlagsForVersion("8.1.3"), flags["8.1.3"])
	assert.Equal(t, FlagsForVersion("8.10.4"), flags["8.10.4"])
}

func TestRunLoadMatrixFailsWhenArtifactMissing(t *testing.T) {
	err := run([]string{"load-matrix", "-artifact", filepath.Join(t.TempDir(), "missing.json")}, io.Discard, io.Discard)
	require.Error(t, err)
}

func TestFlagsByVersion_matchesCheckedInPinnedArtifact(t *testing.T) {
	t.Parallel()

	versions, err := ReadArtifact("../../.github/versions/acceptance-test-matrix.json")
	require.NoError(t, err)
	require.NotEmpty(t, versions)

	flags := FlagsByVersion(versions)
	for _, version := range versions {
		assert.Equal(t, FlagsForVersion(version), flags[version], version)
	}
}

func parseGitHubOutput(t *testing.T, path string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	out := map[string]string{}
	for line := range strings.SplitSeq(string(raw), "\n") {
		if line == "" {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		require.True(t, ok, "invalid GITHUB_OUTPUT line %q", line)
		out[name] = value
	}
	return out
}

func TestRunLoadMatrixPushAddsDedicatedDashboard(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "matrix.json")
	require.NoError(t, WriteArtifact(artifact, []string{"9.4.0"}))
	output := filepath.Join(dir, "output")
	t.Setenv("GITHUB_OUTPUT", output)
	require.NoError(t, run([]string{"load-matrix", "-artifact", artifact, "-event", "push"}, io.Discard, io.Discard))
	assert.JSONEq(t, `[
  {"version":"9.4.0","suite":"general","shard":0,"total_shards":2},
  {"version":"9.4.0","suite":"general","shard":1,"total_shards":2},
  {"version":"9.4.0","suite":"dashboard","shard":0,"total_shards":1}
 ]`, parseGitHubOutput(t, output)["entries"])
}

func TestAcceptanceEntriesDashboardEligibilityUsesNumericVersion(t *testing.T) {
	versions := []string{"8.19.0", "9.3.9-SNAPSHOT", "9.4.0-SNAPSHOT", "9.10.0"}
	entries := acceptanceEntries(versions, "push", []int{0, 1})
	counts := map[string]int{}
	for _, entry := range entries {
		counts[entry.Version]++
		if entry.Suite == "dashboard" {
			assert.Contains(t, []string{"9.4.0-SNAPSHOT", "9.10.0"}, entry.Version)
			assert.Equal(t, 0, entry.Shard)
			assert.Equal(t, 1, entry.TotalShards)
		}
	}
	assert.Equal(t, map[string]int{"8.19.0": 2, "9.3.9-SNAPSHOT": 2, "9.4.0-SNAPSHOT": 3, "9.10.0": 3}, counts)
}

func TestRunLoadMatrixPullRequestPreservesPreparedShards(t *testing.T) {
	for _, shards := range []string{"[0]", "[0,1]"} {
		t.Run(shards, func(t *testing.T) {
			dir := t.TempDir()
			artifact := filepath.Join(dir, "matrix.json")
			require.NoError(t, WriteArtifact(artifact, []string{"9.4.0-SNAPSHOT"}))
			output := filepath.Join(dir, "output")
			t.Setenv("GITHUB_OUTPUT", output)
			require.NoError(t, run([]string{"load-matrix", "-artifact", artifact, "-event", "pull_request", "-shards", shards}, io.Discard, io.Discard))
			var entries []acceptanceMatrixEntry
			require.NoError(t, json.Unmarshal([]byte(parseGitHubOutput(t, output)["entries"]), &entries))
			var indices []int
			require.NoError(t, json.Unmarshal([]byte(shards), &indices))
			require.Len(t, entries, len(indices))
			for i, entry := range entries {
				assert.Equal(t, "general", entry.Suite)
				assert.Equal(t, indices[i], entry.Shard)
				assert.Equal(t, 2, entry.TotalShards)
			}
		})
	}
}

func TestRunLoadMatrixRejectsInvalidPreparedShards(t *testing.T) {
	for _, shards := range []string{"[]", "null", "[-1]", "[2]", "[0,0]", "[1,0]", "[0,1,2]", "[0.5]", "invalid"} {
		t.Run(shards, func(t *testing.T) {
			dir := t.TempDir()
			artifact := filepath.Join(dir, "matrix.json")
			require.NoError(t, WriteArtifact(artifact, []string{"9.4.0"}))
			output := filepath.Join(dir, "output")
			t.Setenv("GITHUB_OUTPUT", output)
			err := run([]string{"load-matrix", "-artifact", artifact, "-event", "pull_request", "-shards", shards}, io.Discard, io.Discard)
			require.ErrorContains(t, err, "invalid shards")
			_, statErr := os.Stat(output)
			assert.True(t, os.IsNotExist(statErr))
		})
	}
}

func TestRunLoadMatrixMergeGroupPreservesGeneralAndDashboardAssignments(t *testing.T) {
	dir := t.TempDir()
	artifact := filepath.Join(dir, "matrix.json")
	versions := []string{"8.19.0", "9.3.9-SNAPSHOT", "9.4.0-SNAPSHOT", "9.10.0"}
	require.NoError(t, WriteArtifact(artifact, versions))
	output := filepath.Join(dir, "output")
	t.Setenv("GITHUB_OUTPUT", output)
	require.NoError(t, run([]string{"load-matrix", "--artifact", artifact, "--event", "merge_group", "--shards", "[0]"}, io.Discard, io.Discard))
	var entries []acceptanceMatrixEntry
	require.NoError(t, json.Unmarshal([]byte(parseGitHubOutput(t, output)["entries"]), &entries))
	assignments := map[string][]acceptanceMatrixEntry{}
	for _, entry := range entries {
		assignments[entry.Version] = append(assignments[entry.Version], entry)
	}
	require.Len(t, assignments, len(versions))
	for _, version := range versions {
		expected := []acceptanceMatrixEntry{
			{Version: version, Suite: "general", Shard: 0, TotalShards: 2},
			{Version: version, Suite: "general", Shard: 1, TotalShards: 2},
		}
		if version == "9.4.0-SNAPSHOT" || version == "9.10.0" {
			expected = append(expected, acceptanceMatrixEntry{Version: version, Suite: "dashboard", Shard: 0, TotalShards: 1})
		}
		assert.Equal(t, expected, assignments[version], version)
	}
}
