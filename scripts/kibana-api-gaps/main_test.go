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
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectResourcesRejectsUnknownNames(t *testing.T) {
	all := []Resource{{Name: "a"}, {Name: "b"}}

	got, err := selectResources(all, []string{"b"})
	require.NoError(t, err)
	assert.Equal(t, []Resource{{Name: "b"}}, got)

	got, err = selectResources(all, []string{"b", "a", "b"})
	require.NoError(t, err)
	assert.Equal(t, all, got, "duplicates collapse and canonical order is kept")

	_, err = selectResources(all, []string{"typo"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "typo")

	got, err = selectResources(all, nil)
	require.NoError(t, err)
	assert.Equal(t, all, got)
}

func TestFindPackageBySuffixRequiresExactlyOneMatch(t *testing.T) {
	mod := loadFixture(t, "testdata/fixture")

	path, err := findPackageBySuffix(mod, "/kbapi")
	require.NoError(t, err)
	assert.Equal(t, fixtureKbapi, path)

	_, err = findPackageBySuffix(mod, "/missing")
	require.Error(t, err)
	_, err = findPackageBySuffix(mod, "s")
	require.Error(t, err)
}

func TestRunRejectsMissingAndUnknownSubcommands(t *testing.T) {
	for _, args := range [][]string{nil, {"bogus"}, {"--repo", "."}} {
		require.Error(t, run(args, io.Discard, io.Discard), args)
	}
}

func TestRunDispatchesRecordToItsCommand(t *testing.T) {
	err := run([]string{"record"}, io.Discard, io.Discard)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--report")
}

func TestStringListCollectsRepeatedValues(t *testing.T) {
	var list stringList
	require.NoError(t, list.Set("a"))
	require.NoError(t, list.Set("b"))
	assert.Equal(t, stringList{"a", "b"}, list)
	assert.Equal(t, "[a b]", list.String())
}

func TestCmdDiscoverRejectsAnUnknownFlag(t *testing.T) {
	require.Error(t, cmdDiscover([]string{"--nope"}, io.Discard, io.Discard))
}

func TestFinishDiscoverReportsErrors(t *testing.T) {
	dir := t.TempDir()
	d := &Discovery{Groups: []Group{gapGroup("r", "a", 1, "f")}}
	bad := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte("{broken"), 0o600))
	blocker := filepath.Join(dir, "file")
	require.NoError(t, os.WriteFile(blocker, nil, 0o600))

	none := filepath.Join(dir, "none")
	report := filepath.Join(dir, "r.json")

	require.Error(t, finishDiscover(d, discoverOptions{memoryPath: bad, reportPath: report}, io.Discard), "malformed memory")
	require.Error(t, finishDiscover(d, discoverOptions{memoryPath: none, reportPath: filepath.Join(blocker, "r.json")}, io.Discard), "unwritable report")
	opts := discoverOptions{memoryPath: none, reportPath: report, summaryPath: filepath.Join(blocker, "s.md")}
	require.Error(t, finishDiscover(d, opts, io.Discard), "unwritable summary")
}
