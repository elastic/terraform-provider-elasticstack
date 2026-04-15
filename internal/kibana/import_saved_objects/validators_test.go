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

package importsavedobjects_test

import (
	"context"
	"strings"
	"testing"

	importsavedobjects "github.com/elastic/terraform-provider-elasticstack/internal/kibana/import_saved_objects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestConfigValidators_Count(t *testing.T) {
	r := importsavedobjects.NewResource().(resource.ResourceWithConfigValidators)
	validators := r.ConfigValidators(context.Background())
	assert.Len(t, validators, 2, "expected exactly 2 config validators")
}

// buildRawConfig constructs a tftypes.Value matching the resource schema with the
// given boolean flag values. Unspecified flags are set to null.
func buildRawConfig(t *testing.T, r resource.Resource, createNewCopies, overwrite, compatibilityMode *bool) tfsdk.Config {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	boolVal := func(b *bool) tftypes.Value {
		if b == nil {
			return tftypes.NewValue(tftypes.Bool, nil)
		}
		return tftypes.NewValue(tftypes.Bool, *b)
	}

	attrTypes := schemaResp.Schema.Type().TerraformType(context.Background())
	errorsType := schemaResp.Schema.Attributes["errors"].GetType().TerraformType(context.Background())
	successResultsType := schemaResp.Schema.Attributes["success_results"].GetType().TerraformType(context.Background())
	kbConnType := schemaResp.Schema.Blocks["kibana_connection"].Type().TerraformType(context.Background())

	rawConfig := tftypes.NewValue(attrTypes, map[string]tftypes.Value{
		"id":                   tftypes.NewValue(tftypes.String, nil),
		"space_id":             tftypes.NewValue(tftypes.String, nil),
		"ignore_import_errors": tftypes.NewValue(tftypes.Bool, nil),
		"create_new_copies":    boolVal(createNewCopies),
		"overwrite":            boolVal(overwrite),
		"compatibility_mode":   boolVal(compatibilityMode),
		"file_contents":        tftypes.NewValue(tftypes.String, "{}"),
		"success":              tftypes.NewValue(tftypes.Bool, nil),
		"success_count":        tftypes.NewValue(tftypes.Number, nil),
		"errors":               tftypes.NewValue(errorsType, nil),
		"success_results":      tftypes.NewValue(successResultsType, nil),
		"kibana_connection":    tftypes.NewValue(kbConnType, nil),
	})

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    rawConfig,
	}
}

func TestConfigValidators_NoConflictWhenFalse(t *testing.T) {
	tr := true
	fa := false
	r := importsavedobjects.NewResource()
	rWithValidators := r.(resource.ResourceWithConfigValidators)
	validators := rWithValidators.ConfigValidators(context.Background())
	require.Len(t, validators, 2)

	// create_new_copies = false, overwrite = false — should NOT error
	config := buildRawConfig(t, r, &fa, &fa, nil)
	req := resource.ValidateConfigRequest{Config: config}
	var allDiags diag.Diagnostics
	for _, v := range validators {
		resp := &resource.ValidateConfigResponse{}
		v.ValidateResource(context.Background(), req, resp)
		allDiags.Append(resp.Diagnostics...)
	}
	assert.False(t, allDiags.HasError(), "validators must not fire when both attributes are false")

	// create_new_copies = false, overwrite = true — should NOT error
	allDiags = nil
	config = buildRawConfig(t, r, &fa, &tr, nil)
	req = resource.ValidateConfigRequest{Config: config}
	for _, v := range validators {
		resp := &resource.ValidateConfigResponse{}
		v.ValidateResource(context.Background(), req, resp)
		allDiags.Append(resp.Diagnostics...)
	}
	assert.False(t, allDiags.HasError(), "validators must not fire when only overwrite is true")
}

func TestConfigValidators_CreateNewCopiesConflictsWithOverwrite(t *testing.T) {
	tr := true
	r := importsavedobjects.NewResource()
	rWithValidators := r.(resource.ResourceWithConfigValidators)
	validators := rWithValidators.ConfigValidators(context.Background())
	require.Len(t, validators, 2)

	config := buildRawConfig(t, r, &tr, &tr, nil)
	req := resource.ValidateConfigRequest{Config: config}

	var allDiags diag.Diagnostics
	for _, v := range validators {
		resp := &resource.ValidateConfigResponse{}
		v.ValidateResource(context.Background(), req, resp)
		allDiags.Append(resp.Diagnostics...)
	}
	require.True(t, allDiags.HasError(), "expected a conflict diagnostic when create_new_copies and overwrite are both true")

	found := false
	for _, d := range allDiags.Errors() {
		if strings.Contains(d.Detail(), "create_new_copies") && strings.Contains(d.Detail(), "overwrite") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected conflict diagnostic to mention both create_new_copies and overwrite")
}

func TestConfigValidators_CreateNewCopiesConflictsWithCompatibilityMode(t *testing.T) {
	tr := true
	r := importsavedobjects.NewResource()
	rWithValidators := r.(resource.ResourceWithConfigValidators)
	validators := rWithValidators.ConfigValidators(context.Background())
	require.Len(t, validators, 2)

	config := buildRawConfig(t, r, &tr, nil, &tr)
	req := resource.ValidateConfigRequest{Config: config}

	var allDiags diag.Diagnostics
	for _, v := range validators {
		resp := &resource.ValidateConfigResponse{}
		v.ValidateResource(context.Background(), req, resp)
		allDiags.Append(resp.Diagnostics...)
	}
	require.True(t, allDiags.HasError(), "expected a conflict diagnostic when create_new_copies and compatibility_mode are both true")

	found := false
	for _, d := range allDiags.Errors() {
		if strings.Contains(d.Detail(), "create_new_copies") && strings.Contains(d.Detail(), "compatibility_mode") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected conflict diagnostic to mention both create_new_copies and compatibility_mode")
}
