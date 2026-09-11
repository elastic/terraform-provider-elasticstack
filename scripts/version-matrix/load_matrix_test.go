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

func TestLoadMatrix_matchesCheckedInPinnedArtifact(t *testing.T) {
	t.Parallel()

	versions, err := ReadArtifact("../../.github/versions/acceptance-test-matrix.json")
	require.NoError(t, err)
	require.NotEmpty(t, versions)

	got, flags := LoadMatrix(versions)
	assert.Equal(t, versions, got)
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
