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

package index

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
)

func Test_stringIsJSONObject(t *testing.T) {
	tests := []struct {
		name                  string
		fieldVal              any
		expectedErrsToContain []string
	}{
		{
			name:     "should not return an error for a valid json object",
			fieldVal: "{}",
		},
		{
			name:     "should not return an error for a null",
			fieldVal: "null",
		},

		{
			name:     "should return an error if the field is not a string",
			fieldVal: true,
			expectedErrsToContain: []string{
				"expected type of field-name to be string",
			},
		},
		{
			name:     "should return an error if the field is valid json, but not an object",
			fieldVal: "[]",
			expectedErrsToContain: []string{
				"expected field-name to be a JSON object. Check the documentation for the expected format.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := stringIsJSONObject(tt.fieldVal, "field-name")
			require.Len(t, errors, len(tt.expectedErrsToContain))
			for i, err := range errors {
				require.ErrorContains(t, err, tt.expectedErrsToContain[i])
			}
		})
	}
}

func Test_validateDataStreamOptionsVersion(t *testing.T) {
	dso := &models.Template{
		DataStreamOptions: &models.DataStreamOptions{
			FailureStore: &models.FailureStoreOptions{Enabled: true},
		},
	}
	noOpts := &models.Template{}

	tests := []struct {
		name        string
		serverVer   string
		templ       *models.Template
		wantErr     bool
		errContains string
	}{
		{
			name:        "below minimum version with data_stream_options configured",
			serverVer:   "9.0.0",
			templ:       dso,
			wantErr:     true,
			errContains: "9.1.0",
		},
		{
			name:      "at minimum version",
			serverVer: "9.1.0",
			templ:     dso,
			wantErr:   false,
		},
		{
			name:      "above minimum version",
			serverVer: "9.2.0",
			templ:     dso,
			wantErr:   false,
		},
		{
			name:      "below minimum version but no data_stream_options",
			serverVer: "9.0.0",
			templ:     noOpts,
			wantErr:   false,
		},
		{
			name:      "nil template",
			serverVer: "9.0.0",
			templ:     nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := version.Must(version.NewVersion(tt.serverVer))
			diags := validateDataStreamOptionsVersion(v, tt.templ)
			if tt.wantErr {
				require.True(t, diags.HasError(), "expected error diagnostic")
				require.Contains(t, diags[0].Summary, tt.errContains)
			} else {
				require.False(t, diags.HasError(), "unexpected error: %v", diags)
			}
		})
	}
}
