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

package utils_test

import (
	"context"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/alerting_v2/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func validateString(t *testing.T, v validator.String, value types.String) validator.StringResponse {
	t.Helper()
	resp := validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{
		Path:        path.Root("attr"),
		ConfigValue: value,
	}, &resp)
	return resp
}

func TestEntityIDValidator_acceptsValidIDs(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"cpu-high_prod", "a", "ABC-123_xyz", strings.Repeat("x", 150)} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.EntityIDValidator(), types.StringValue(id))
			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestEntityIDValidator_rejectsInvalidIDs(t *testing.T) {
	t.Parallel()

	for name, id := range map[string]string{
		"space and punctuation": "my rule!",
		"empty":                 "",
		"too long":              strings.Repeat("x", 151),
		"dot":                   "a.b",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.EntityIDValidator(), types.StringValue(id))
			require.True(t, resp.Diagnostics.HasError())
			detail := resp.Diagnostics[0].Detail()
			require.Contains(t, detail, "1-150")
			require.Contains(t, detail, "letters, digits, underscores, or hyphens")
		})
	}
}

func TestDurationValidator_acceptsEveryUnit(t *testing.T) {
	t.Parallel()

	for _, d := range []string{"500ms", "5s", "1m", "12h", "365d", "52w"} {
		t.Run(d, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidator(), types.StringValue(d))
			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}
}

func TestDurationValidator_rejectsMalformedValues(t *testing.T) {
	t.Parallel()

	for name, d := range map[string]string{
		"spelled out unit": "5 minutes",
		"year unit":        "1y",
		"no unit":          "5",
		"negative":         "-5s",
		"empty":            "",
		"uppercase unit":   "5M",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidator(), types.StringValue(d))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "ms, s, m, h, d, or w")
		})
	}
}

func TestDurationValidator_enforces365DayCeiling(t *testing.T) {
	t.Parallel()

	accepted := []string{"365d", "8760h", "525600m", "31536000s", "31536000000ms", "52w"}
	for _, d := range accepted {
		t.Run("accepts "+d, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidator(), types.StringValue(d))
			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}

	rejected := []string{"366d", "8761h", "525601m", "31536001s", "31536000001ms", "53w"}
	for _, d := range rejected {
		t.Run("rejects "+d, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidator(), types.StringValue(d))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "365d")
		})
	}
}

func TestDurationValidatorAtLeast_enforcesMinimum(t *testing.T) {
	t.Parallel()

	for _, d := range []string{"5s", "5000ms", "1m", "365d"} {
		t.Run("accepts "+d, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidatorAtLeast("5s"), types.StringValue(d))
			require.False(t, resp.Diagnostics.HasError(), "unexpected diagnostics: %v", resp.Diagnostics)
		})
	}

	for _, d := range []string{"1s", "4999ms", "0m"} {
		t.Run("rejects "+d, func(t *testing.T) {
			t.Parallel()
			resp := validateString(t, utils.DurationValidatorAtLeast("5s"), types.StringValue(d))
			require.True(t, resp.Diagnostics.HasError())
			require.Contains(t, resp.Diagnostics[0].Detail(), "5s")
		})
	}

	t.Run("still rejects values above 365d", func(t *testing.T) {
		t.Parallel()
		resp := validateString(t, utils.DurationValidatorAtLeast("5s"), types.StringValue("366d"))
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("unknown values are skipped", func(t *testing.T) {
		t.Parallel()
		resp := validateString(t, utils.DurationValidatorAtLeast("5s"), types.StringUnknown())
		require.False(t, resp.Diagnostics.HasError())
	})
}
