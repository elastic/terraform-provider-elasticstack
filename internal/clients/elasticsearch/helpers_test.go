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

package elasticsearch

import (
	"io"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unmarshalableBody is a type that cannot be marshalled to JSON because it
// contains a channel field (channels have no JSON representation).
type unmarshalableBody struct {
	Ch chan int
}

func TestDoFWWriteMarshalError(t *testing.T) {
	tests := []struct {
		name          string
		marshalErrMsg string
	}{
		{
			name:          "custom marshal error message is surfaced",
			marshalErrMsg: "Unable to marshal my resource",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			diags := doFWWrite(
				nil, // apiClient — never reached when marshal fails
				unmarshalableBody{Ch: make(chan int)},
				tc.marshalErrMsg,
				"call error",
				"response error",
				func(_ *elasticsearch.Client, _ io.Reader) (*esapi.Response, error) {
					called = true
					return nil, nil
				},
			)

			require.True(t, diags.HasError(), "expected an error diagnostic")
			assert.False(t, called, "fn must not be called when marshal fails")
			assert.Equal(t, tc.marshalErrMsg, diags[0].Summary())
		})
	}
}

func TestDoSDKWriteMarshalError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "marshal error is surfaced as SDK diagnostic"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			diags := doSDKWrite(
				nil, // apiClient — never reached when marshal fails
				unmarshalableBody{Ch: make(chan int)},
				"response error",
				func(_ *elasticsearch.Client, _ io.Reader) (*esapi.Response, error) {
					called = true
					return nil, nil
				},
			)

			require.True(t, diags.HasError(), "expected an error diagnostic")
			assert.False(t, called, "fn must not be called when marshal fails")
			assert.NotEmpty(t, diags[0].Summary, "error summary should not be empty")
		})
	}
}
