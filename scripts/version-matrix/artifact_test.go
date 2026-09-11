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

func TestReadArtifactReturnsPinnedContents(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, os.WriteFile(path, []byte("[\n  \"8.19.21\",\n  \"9.6.0-SNAPSHOT\"\n]\n"), 0o644))

	got, err := ReadArtifact(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.19.21", "9.6.0-SNAPSHOT"}, got)
}

func TestWriteArtifactSortsAscendingWithSnapshotLast(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(path, []string{"9.0.8", "8.19.21-SNAPSHOT", "8.10.4"}))

	got, err := ReadArtifact(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.10.4", "9.0.8", "8.19.21-SNAPSHOT"}, got)
}

func TestWriteArtifactSortsTwoDigitMinorsBeforeLaterPatches(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "acceptance-test-matrix.json")
	require.NoError(t, WriteArtifact(path, []string{"8.19.21", "8.1.3", "8.10.4", "9.0.8-SNAPSHOT"}))

	got, err := ReadArtifact(path)
	require.NoError(t, err)
	assert.Equal(t, []string{"8.1.3", "8.10.4", "8.19.21", "9.0.8-SNAPSHOT"}, got)
}

func TestVersionsEqualDetectsNoOpRegardlessOfOrder(t *testing.T) {
	t.Parallel()

	assert.True(t, VersionsEqual(
		[]string{"9.6.0-SNAPSHOT", "8.19.21"},
		[]string{"8.19.21", "9.6.0-SNAPSHOT"},
	))
}

func TestVersionsEqualDetectsChange(t *testing.T) {
	t.Parallel()

	assert.False(t, VersionsEqual([]string{"8.19.17"}, []string{"8.19.21"}))
}
