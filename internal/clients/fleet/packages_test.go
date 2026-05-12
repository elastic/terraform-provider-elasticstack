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

package fleet

import (
	"testing"
)

func TestIsManifestYAML(t *testing.T) {
	tests := []struct {
		name      string
		entryName string
		want      bool
	}{
		{"bare filename", "manifest.yml", true},
		{"top-level dir prefix", "mypackage-1.0.0/manifest.yml", true},
		{"nested", "mypackage/subdir/manifest.yml", true},
		{"nested manifest under data_stream", "mypackage-1.0.0/data_stream/logs/manifest.yml", true},
		{"different file", "mypackage-1.0.0/README.md", false},
		{"different yaml extension", "mypackage-1.0.0/manifest.yaml", false},
		{"empty string", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isManifestYAML(tc.entryName); got != tc.want {
				t.Errorf("isManifestYAML(%q) = %v, want %v", tc.entryName, got, tc.want)
			}
		})
	}
}

func TestExtractPackageNameVersion(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantVersion string
	}{
		{
			name:        "plain values",
			content:     "name: my-package\nversion: 1.2.3\n",
			wantName:    "my-package",
			wantVersion: "1.2.3",
		},
		{
			name:        "quoted version",
			content:     "name: my-package\nversion: \"1.2.3\"\n",
			wantName:    "my-package",
			wantVersion: "1.2.3",
		},
		{
			name:        "single-quoted version",
			content:     "name: my-package\nversion: '1.0.0'\n",
			wantName:    "my-package",
			wantVersion: "1.0.0",
		},
		{
			name:        "name only",
			content:     "name: only-name\n",
			wantName:    "only-name",
			wantVersion: "",
		},
		{
			name:        "no name field",
			content:     "version: 1.0.0\nformat_version: 1.0.0\n",
			wantName:    "",
			wantVersion: "1.0.0",
		},
		{
			name:        "empty content",
			content:     "",
			wantName:    "",
			wantVersion: "",
		},
		{
			name:        "extra whitespace after colon",
			content:     "name:   my-package\nversion:   2.0.0\n",
			wantName:    "my-package",
			wantVersion: "2.0.0",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotVersion := extractPackageNameVersion([]byte(tc.content))
			if gotName != tc.wantName {
				t.Errorf("name = %q, want %q", gotName, tc.wantName)
			}
			if gotVersion != tc.wantVersion {
				t.Errorf("version = %q, want %q", gotVersion, tc.wantVersion)
			}
		})
	}
}
