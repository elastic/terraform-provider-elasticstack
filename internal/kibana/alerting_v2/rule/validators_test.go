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

package rule

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func validateArtifactData(t *testing.T, value types.String) validator.StringResponse {
	t.Helper()
	resp := validator.StringResponse{}
	artifactDataValidator().ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("artifacts").AtListIndex(0).AtName("data"),
		ConfigValue: value,
	}, &resp)
	return resp
}

func jsonObjectWithKeys(count int, keyPrefix string) string {
	keys := make([]string, 0, count)
	for i := range count {
		keys = append(keys, `"`+keyPrefix+strconv.Itoa(i)+`": 1`)
	}
	return "{" + strings.Join(keys, ",") + "}"
}

func TestArtifactDataValidator_rejectsNonObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{
		"array":    `["a"]`,
		"string":   `"a"`,
		"not json": "not json",
		"null":     "null",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateArtifactData(t, types.StringValue(value))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Summary()+" "+resp.Diagnostics[0].Detail(), "JSON object")
		})
	}
}

func TestArtifactDataValidator_enforcesTopLevelKeyCount(t *testing.T) {
	t.Parallel()

	t.Run("accepts exactly 32 keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(jsonObjectWithKeys(32, "k")))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	t.Run("rejects 33 keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(jsonObjectWithKeys(33, "k")))
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics[0].Detail(), "at most 32 top-level keys")
	})
}

func TestArtifactDataValidator_enforcesKeyLength(t *testing.T) {
	t.Parallel()

	t.Run("accepts a 256-character key and unbounded nested keys", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringValue(`{"`+strings.Repeat("k", 256)+`": 1, "a": {"`+strings.Repeat("n", 300)+`": true}}`))
		require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
	})

	for name, value := range map[string]string{
		"empty key":    `{"": 1}`,
		"too long key": `{"` + strings.Repeat("k", 257) + `": 1}`,
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			t.Parallel()
			resp := validateArtifactData(t, types.StringValue(value))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "1-256 characters")
		})
	}

	t.Run("unknown values are skipped", func(t *testing.T) {
		t.Parallel()
		resp := validateArtifactData(t, types.StringUnknown())
		require.False(t, resp.Diagnostics.HasError())
	})
}
