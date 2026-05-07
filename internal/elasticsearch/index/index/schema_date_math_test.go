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
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateName runs all name validators configured in the index resource schema
// against the given value and reports whether any error diagnostics were produced.
func validateName(t *testing.T, value string) (hasError bool) {
	t.Helper()
	s := getSchema(context.Background())
	nameAttrRaw, ok := s.Attributes["name"]
	require.True(t, ok, "name attribute not found in schema")

	nameAttr, ok := nameAttrRaw.(schema.StringAttribute)
	require.True(t, ok, "name attribute is not a schema.StringAttribute")

	req := validator.StringRequest{
		Path:           path.Root("name"),
		PathExpression: path.MatchRoot("name"),
		ConfigValue:    types.StringValue(value),
	}

	for _, v := range nameAttr.Validators {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), req, resp)
		if resp.Diagnostics.HasError() {
			return true
		}
	}
	return false
}

// TestIndexNameValidation_StaticNames verifies that static index names that satisfy
// the existing lowercase-name rules are accepted.
func TestIndexNameValidation_StaticNames(t *testing.T) {
	validStatic := []string{
		"my-index",
		"logs-2024.01.15",
		"a",
		"abc123",
		"log.data",
		"test_index",
	}
	for _, name := range validStatic {
		t.Run("valid static: "+name, func(t *testing.T) {
			assert.False(t, validateName(t, name), "expected static name %q to be valid", name)
		})
	}
}

// TestIndexNameValidation_DateMathNames verifies that plain date math expressions
// with proper angle-bracket wrapping and at least one {…} section are accepted.
func TestIndexNameValidation_DateMathNames(t *testing.T) {
	validDateMath := []string{
		`<logs-{now/d}>`,
		`<logs-{now/M}>`,
		`<logs-{now/d{yyyy.MM.dd}}>`,
	}
	for _, name := range validDateMath {
		t.Run("valid date math: "+name, func(t *testing.T) {
			assert.False(t, validateName(t, name), "expected date math name %q to be valid", name)
		})
	}
}

// TestIndexNameValidation_InvalidInputs verifies that values satisfying neither the
// static-name validator nor the date-math validator are rejected.
func TestIndexNameValidation_InvalidInputs(t *testing.T) {
	invalid := []struct {
		label string
		name  string
	}{
		{"starts with dash", "-myindex"},
		{"starts with underscore", "_myindex"},
		{"starts with plus", "+myindex"},
		{"uppercase", "MyIndex"},
		{"dot", "."},
		{"double dot", ".."},
		{"angle brackets no braces", "<logs>"},
		{"braces but no angle brackets", "logs-{now/d}"},
		{"suffix after date math section", `<prefix-{now/d}-suffix>`},
	}
	for _, tc := range invalid {
		t.Run("invalid: "+tc.label, func(t *testing.T) {
			assert.True(t, validateName(t, tc.name), "expected name %q to be invalid", tc.name)
		})
	}
}
