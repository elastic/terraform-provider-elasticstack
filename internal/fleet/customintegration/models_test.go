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

package customintegration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want string
	}{
		{"plain .zip", "my-package.zip", "application/zip"},
		{".ZIP uppercase", "MY.ZIP", "application/zip"},
		{".tar.gz", "my-package.tar.gz", "application/gzip"},
		{".tgz", "my-package.tgz", "application/gzip"},
		{".gz", "my-package.gz", "application/gzip"},
		{"no extension defaults to zip", "packagefile", "application/zip"},
		{"path with directory", "/tmp/builds/foo-1.0.0.zip", "application/zip"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, detectContentType(tc.path))
		})
	}
}

func TestSha256File(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	f1 := filepath.Join(dir, "a")
	f2 := filepath.Join(dir, "b")
	f3 := filepath.Join(dir, "c")
	require.NoError(t, os.WriteFile(f1, []byte("hello world"), 0o600))
	require.NoError(t, os.WriteFile(f2, []byte("hello world"), 0o600))
	require.NoError(t, os.WriteFile(f3, []byte("hello world!"), 0o600))

	h1, err := sha256File(f1)
	require.NoError(t, err)
	h2, err := sha256File(f2)
	require.NoError(t, err)
	h3, err := sha256File(f3)
	require.NoError(t, err)

	assert.Equal(t, h1, h2, "identical content must produce identical hashes")
	assert.NotEqual(t, h1, h3, "different content must produce different hashes")
	// SHA-256 of "hello world" is a well-known value; the helper must
	// return the standard lowercase-hex form so downstream checksum
	// comparisons are stable across platforms.
	assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", h1)
}

func TestSha256File_Missing(t *testing.T) {
	t.Parallel()
	_, err := sha256File(filepath.Join(t.TempDir(), "does-not-exist"))
	require.Error(t, err)
}

func TestGetPackageID_IsDeterministic(t *testing.T) {
	t.Parallel()

	a := getPackageID("my_pkg", "1.2.3")
	b := getPackageID("my_pkg", "1.2.3")
	c := getPackageID("my_pkg", "1.2.4")

	assert.NotEmpty(t, a)
	assert.Equal(t, a, b, "same (name, version) must produce identical IDs")
	assert.NotEqual(t, a, c, "different versions must produce different IDs")
}

func TestPickInstalledVersion(t *testing.T) {
	t.Parallel()

	installedStr := "installed"
	notInstalledStr := "not_installed"
	installed := &installedStr
	notInstalled := &notInstalledStr

	cases := []struct {
		name string
		in   []kbapi.PackageListItem
		pkg  string
		want string
	}{
		{
			name: "single installed entry",
			in:   []kbapi.PackageListItem{{Name: "pkg_a", Version: "1.0.0", Status: installed}},
			pkg:  "pkg_a",
			want: "1.0.0",
		},
		{
			name: "ignores entries with a different name",
			in: []kbapi.PackageListItem{
				{Name: "pkg_b", Version: "9.9.9", Status: installed},
				{Name: "pkg_a", Version: "0.1.0", Status: installed},
			},
			pkg:  "pkg_a",
			want: "0.1.0",
		},
		{
			name: "picks highest semver among installed",
			in: []kbapi.PackageListItem{
				{Name: "pkg_a", Version: "1.0.0", Status: installed},
				{Name: "pkg_a", Version: "1.5.2", Status: installed},
				{Name: "pkg_a", Version: "0.9.0", Status: installed},
			},
			pkg:  "pkg_a",
			want: "1.5.2",
		},
		{
			name: "ignores not-installed entries",
			in: []kbapi.PackageListItem{
				{Name: "pkg_a", Version: "2.0.0", Status: notInstalled},
				{Name: "pkg_a", Version: "1.0.0", Status: installed},
			},
			pkg:  "pkg_a",
			want: "1.0.0",
		},
		{
			name: "no matching package returns empty",
			in:   []kbapi.PackageListItem{{Name: "other", Version: "1.0.0", Status: installed}},
			pkg:  "pkg_a",
			want: "",
		},
		{
			name: "nil status treated as installed",
			in:   []kbapi.PackageListItem{{Name: "pkg_a", Version: "3.0.0", Status: nil}},
			pkg:  "pkg_a",
			want: "3.0.0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pickInstalledVersion(tc.in, tc.pkg))
		})
	}
}
