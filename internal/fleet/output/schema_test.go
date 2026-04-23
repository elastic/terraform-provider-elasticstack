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

package output

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaSSLVerificationModeValidator(t *testing.T) {
	t.Parallel()

	s := getSchema()

	sslAttr, ok := s.Attributes["ssl"].(schema.SingleNestedAttribute)
	require.True(t, ok, "expected ssl to be a SingleNestedAttribute")

	vmAttr, ok := sslAttr.Attributes["verification_mode"].(schema.StringAttribute)
	require.True(t, ok, "expected verification_mode to be a StringAttribute")
	require.NotEmpty(t, vmAttr.Validators, "expected verification_mode to have validators")

	validValues := []string{"certificate", "full", "none", "strict"}

	// Valid values should pass all validators without errors.
	for _, v := range validValues {
		t.Run("valid/"+v, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{ConfigValue: types.StringValue(v)}
			var resp validator.StringResponse
			for _, val := range vmAttr.Validators {
				val.ValidateString(context.Background(), req, &resp)
			}
			assert.False(t, resp.Diagnostics.HasError(), "expected %q to be valid, got: %v", v, resp.Diagnostics)
		})
	}

	// Invalid value should fail validation.
	t.Run("invalid/value", func(t *testing.T) {
		t.Parallel()
		req := validator.StringRequest{ConfigValue: types.StringValue("invalid")}
		var resp validator.StringResponse
		for _, val := range vmAttr.Validators {
			val.ValidateString(context.Background(), req, &resp)
		}
		assert.True(t, resp.Diagnostics.HasError(), "expected \"invalid\" to fail validation")
	})
}

func TestSchemaIncludesRemoteElasticsearchTypeAndServiceToken(t *testing.T) {
	t.Parallel()

	s := getSchema()

	typeAttr, ok := s.Attributes["type"].(schema.StringAttribute)
	require.True(t, ok)
	require.NotEmpty(t, typeAttr.Validators)

	allowedType := false
	for _, validator := range typeAttr.Validators {
		if strings.Contains(validator.Description(context.Background()), "remote_elasticsearch") {
			allowedType = true
			break
		}
	}
	assert.True(t, allowedType, "expected remote_elasticsearch to be an allowed type")

	serviceTokenAttr, ok := s.Attributes["service_token"].(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, serviceTokenAttr.Sensitive)
	assert.True(t, serviceTokenAttr.Optional)
	assert.NotEmpty(t, serviceTokenAttr.Validators)
}
