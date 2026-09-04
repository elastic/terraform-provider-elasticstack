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

package fileutil_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/fileutil"
	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, body string) string {
	t.Helper()
	fpath := filepath.Join(t.TempDir(), "content")
	require.NoError(t, os.WriteFile(fpath, []byte(body), 0o600))
	return fpath
}

func TestSHA256HexDigest(t *testing.T) {
	fpath := writeTempFile(t, "hello world")
	sum := sha256.Sum256([]byte("hello world"))
	want := hex.EncodeToString(sum[:])

	got, err := fileutil.SHA256HexDigest(fpath)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestSHA256HexDigest_missingFile(t *testing.T) {
	_, err := fileutil.SHA256HexDigest(filepath.Join(t.TempDir(), "missing"))
	require.Error(t, err)
}

func TestFileChecksumDrifted_noPriorStateAlwaysChanged(t *testing.T) {
	fpath := writeTempFile(t, "body")

	newChecksum, changed, err := fileutil.FileChecksumDrifted(fpath, "", false)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotEmpty(t, newChecksum)
}

func TestFileChecksumDrifted_unchangedWhenChecksumMatches(t *testing.T) {
	fpath := writeTempFile(t, "stable body")

	prior, err := fileutil.SHA256HexDigest(fpath)
	require.NoError(t, err)

	newChecksum, changed, err := fileutil.FileChecksumDrifted(fpath, prior, true)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, prior, newChecksum)
}

func TestFileChecksumDrifted_changedWhenFileDiffersFromState(t *testing.T) {
	fpath := writeTempFile(t, "new body")

	_, changed, err := fileutil.FileChecksumDrifted(fpath, "stale-checksum", true)
	require.NoError(t, err)
	require.True(t, changed)
}

func TestFileChecksumDrifted_missingFileErrors(t *testing.T) {
	_, changed, err := fileutil.FileChecksumDrifted(filepath.Join(t.TempDir(), "missing"), "", false)
	require.Error(t, err)
	require.False(t, changed)
}
