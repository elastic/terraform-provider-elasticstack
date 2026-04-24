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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseUploadPackageResponse covers both response shapes the Fleet API
// has served for POST /api/fleet/epm/packages (the modern `items` envelope
// and the legacy `response` envelope), along with malformed payloads. The
// helper underpins the UploadPackage wrapper's name/version extraction.
func TestParseUploadPackageResponse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		body        string
		wantName    string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "modern items envelope with version",
			body:        `{"_meta":{"name":"my_pkg"},"items":[{"id":"asset-1","type":"ingest_pipeline","version":"1.2.3"}]}`,
			wantName:    "my_pkg",
			wantVersion: "1.2.3",
		},
		{
			name:        "items envelope without version",
			body:        `{"_meta":{"name":"my_pkg"},"items":[{"id":"asset-1","type":"ingest_pipeline"}]}`,
			wantName:    "my_pkg",
			wantVersion: "",
		},
		{
			name:        "legacy response envelope",
			body:        `{"_meta":{"name":"older_pkg"},"response":[{"id":"asset-1","type":"dashboard","version":"0.1.0"}]}`,
			wantName:    "older_pkg",
			wantVersion: "0.1.0",
		},
		{
			name:        "mixed: items present but version in response",
			body:        `{"_meta":{"name":"mixed_pkg"},"items":[{"id":"asset-1","type":"dashboard"}],"response":[{"id":"a","type":"b","version":"9.9.9"}]}`,
			wantName:    "mixed_pkg",
			wantVersion: "9.9.9",
		},
		{
			name:     "empty envelope",
			body:     `{}`,
			wantName: "",
		},
		{
			name:    "invalid json",
			body:    `not json`,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, version, err := parseUploadPackageResponse([]byte(tc.body))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, name)
			assert.Equal(t, tc.wantVersion, version)
		})
	}
}
