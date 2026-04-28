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

package kibana

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

// TestConfiguredString pins the behaviour that fixes
// https://github.com/elastic/terraform-provider-elasticstack/issues/1881:
// the helper must treat an explicitly-set empty string as "user configured
// this to empty" (so it appears in the outbound request body and Kibana
// clears the field), and only return nil when the attribute is genuinely
// absent or null in configuration.
func TestConfiguredString(t *testing.T) {
	t.Parallel()

	objType := cty.Object(map[string]cty.Type{
		"description": cty.String,
		"image_url":   cty.String,
		"initials":    cty.String,
	})

	tests := []struct {
		name      string
		raw       cty.Value
		attr      string
		wantNil   bool
		wantValue string
	}{
		{
			name:    "null root returns nil",
			raw:     cty.NullVal(objType),
			attr:    "description",
			wantNil: true,
		},
		{
			name:    "attribute missing from type returns nil",
			raw:     cty.ObjectVal(map[string]cty.Value{"description": cty.StringVal("x"), "image_url": cty.StringVal("y"), "initials": cty.StringVal("z")}),
			attr:    "nonexistent",
			wantNil: true,
		},
		{
			name:    "null attribute returns nil",
			raw:     cty.ObjectVal(map[string]cty.Value{"description": cty.NullVal(cty.String), "image_url": cty.NullVal(cty.String), "initials": cty.NullVal(cty.String)}),
			attr:    "description",
			wantNil: true,
		},
		{
			name:      "explicit empty string returns pointer to empty string (the bug fix)",
			raw:       cty.ObjectVal(map[string]cty.Value{"description": cty.StringVal(""), "image_url": cty.NullVal(cty.String), "initials": cty.NullVal(cty.String)}),
			attr:      "description",
			wantNil:   false,
			wantValue: "",
		},
		{
			name:      "explicit value returns pointer to value",
			raw:       cty.ObjectVal(map[string]cty.Value{"description": cty.StringVal("hello world"), "image_url": cty.NullVal(cty.String), "initials": cty.NullVal(cty.String)}),
			attr:      "description",
			wantNil:   false,
			wantValue: "hello world",
		},
		{
			name:      "distinct attributes don't bleed — image_url reset while description kept",
			raw:       cty.ObjectVal(map[string]cty.Value{"description": cty.StringVal("kept"), "image_url": cty.StringVal(""), "initials": cty.NullVal(cty.String)}),
			attr:      "image_url",
			wantNil:   false,
			wantValue: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := configuredString(tc.raw, tc.attr)
			if tc.wantNil {
				assert.Nil(t, got, "expected nil, got pointer to %q", safeDeref(got))
			} else {
				assert.NotNil(t, got, "expected pointer to %q, got nil", tc.wantValue)
				if got != nil {
					assert.Equal(t, tc.wantValue, *got)
				}
			}
		})
	}
}

func safeDeref(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}
