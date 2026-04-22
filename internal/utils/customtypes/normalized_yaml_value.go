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
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gopkg.in/yaml.v3"
)

var (
	_ basetypes.StringValuable                   = NormalizedYamlValue{}
	_ basetypes.StringValuableWithSemanticEquals = NormalizedYamlValue{}
	_ xattr.ValidateableAttribute                = NormalizedYamlValue{}
)

// NormalizedYamlValue is a custom value type for YAML attributes.
type NormalizedYamlValue struct {
	basetypes.StringValue
}

// NewNormalizedYamlNull creates a NormalizedYamlValue with a null value.
func NewNormalizedYamlNull() NormalizedYamlValue {
	return NormalizedYamlValue{StringValue: basetypes.NewStringNull()}
}

// NewNormalizedYamlUnknown creates a NormalizedYamlValue with an unknown value.
func NewNormalizedYamlUnknown() NormalizedYamlValue {
	return NormalizedYamlValue{StringValue: basetypes.NewStringUnknown()}
}

// NewNormalizedYamlValue creates a NormalizedYamlValue with a known value.
func NewNormalizedYamlValue(value string) NormalizedYamlValue {
	return NormalizedYamlValue{StringValue: basetypes.NewStringValue(value)}
}

// Type returns the type of the value.
func (v NormalizedYamlValue) Type(_ context.Context) attr.Type {
	return NormalizedYamlType{}
}

// ValidateAttribute validates that the string value is valid YAML.
func (v NormalizedYamlValue) ValidateAttribute(_ context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	var out any
	if err := yaml.Unmarshal([]byte(v.ValueString()), &out); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid YAML",
			fmt.Sprintf("A string value was provided that is not valid YAML.\n\nPath: %s\nError: %s", req.Path, err),
		)
	}
}

// StringSemanticEquals returns true if both values are semantically equal YAML
// (i.e. parse to the same structure), regardless of whitespace or key ordering.
func (v NormalizedYamlValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(NormalizedYamlValue)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)
		return false, diags
	}

	if v.IsNull() {
		return newValue.IsNull(), diags
	}

	if v.IsUnknown() {
		return newValue.IsUnknown(), diags
	}

	thisJSON, err := yamlToCanonicalJSON(v.ValueString())
	if err != nil {
		// Not valid YAML — fall back to string equality; ValidateAttribute will surface the error
		return v.Equal(newValue.StringValue), diags
	}

	thatJSON, err := yamlToCanonicalJSON(newValue.ValueString())
	if err != nil {
		return v.Equal(newValue.StringValue), diags
	}

	return thisJSON == thatJSON, diags
}

// yamlToCanonicalJSON parses YAML and re-encodes it as canonical (sorted-key) JSON.
func yamlToCanonicalJSON(yamlStr string) (string, error) {
	var parsed any
	if err := yaml.Unmarshal([]byte(yamlStr), &parsed); err != nil {
		return "", err
	}

	b, err := json.Marshal(normalizeYamlValue(parsed))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// normalizeYamlValue recursively converts yaml.v3 map types to map[string]any
// so that encoding/json marshals them with sorted keys.
func normalizeYamlValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = normalizeYamlValue(v2)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = normalizeYamlValue(v2)
		}
		return out
	default:
		return val
	}
}
