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

package customtypes

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizedYamlValue_StringSemanticEquals(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "identical YAML",
			a:        "name: foo\nenabled: true\n",
			b:        "name: foo\nenabled: true\n",
			expected: true,
		},
		{
			name:     "different key order",
			a:        "enabled: true\nname: foo\n",
			b:        "name: foo\nenabled: true\n",
			expected: true,
		},
		{
			name:     "extra blank lines",
			a:        "name: foo\n\nenabled: true\n",
			b:        "name: foo\nenabled: true\n",
			expected: true,
		},
		{
			name:     "different values",
			a:        "name: foo\nenabled: true\n",
			b:        "name: bar\nenabled: true\n",
			expected: false,
		},
		{
			name:     "array order matters",
			a:        "steps:\n  - name: a\n  - name: b\n",
			b:        "steps:\n  - name: b\n  - name: a\n",
			expected: false,
		},
		{
			name:     "nested keys reordered",
			a:        "config:\n  b: 2\n  a: 1\n",
			b:        "config:\n  a: 1\n  b: 2\n",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewNormalizedYamlValue(tt.a)
			b := NewNormalizedYamlValue(tt.b)

			equal, diags := a.StringSemanticEquals(ctx, b)
			require.False(t, diags.HasError())
			assert.Equal(t, tt.expected, equal)
		})
	}
}

func TestNormalizedYamlValue_ValidateAttribute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid YAML",
			value:   "name: foo\nenabled: true\n",
			wantErr: false,
		},
		{
			name:    "invalid YAML",
			value:   "name: :\n  broken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewNormalizedYamlValue(tt.value)
			req := xattr.ValidateAttributeRequest{Path: path.Root("configuration_yaml")}
			resp := &xattr.ValidateAttributeResponse{}
			v.ValidateAttribute(ctx, req, resp)
			assert.Equal(t, tt.wantErr, resp.Diagnostics.HasError())
		})
	}
}

func TestNormalizedYamlValue_NullAndUnknown(t *testing.T) {
	ctx := context.Background()

	null := NewNormalizedYamlNull()
	unknown := NewNormalizedYamlUnknown()
	known := NewNormalizedYamlValue("name: foo\n")

	equal, diags := null.StringSemanticEquals(ctx, NewNormalizedYamlNull())
	require.False(t, diags.HasError())
	assert.True(t, equal)

	equal, diags = null.StringSemanticEquals(ctx, known)
	require.False(t, diags.HasError())
	assert.False(t, equal)

	equal, diags = unknown.StringSemanticEquals(ctx, NewNormalizedYamlUnknown())
	require.False(t, diags.HasError())
	assert.True(t, equal)

	// ValidateAttribute on null/unknown is a no-op
	req := xattr.ValidateAttributeRequest{Path: path.Root("test")}
	resp := &xattr.ValidateAttributeResponse{}
	null.ValidateAttribute(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError())
}
