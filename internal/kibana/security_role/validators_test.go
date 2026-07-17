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

package security_role

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateResourceKibanaPrivileges(t *testing.T) {
	t.Parallel()

	featureElemType := types.ObjectType{AttrTypes: kibanaFeatureAttrTypes()}
	spacesSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("*")})

	featureObj := types.ObjectValueMust(kibanaFeatureAttrTypes(), map[string]attr.Value{
		attrName:       types.StringValue("discover"),
		attrPrivileges: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("read")}),
	})
	knownFeatureSet := types.SetValueMust(featureElemType, []attr.Value{featureObj})
	knownBaseSet := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("all")})
	emptyFeatureSet := types.SetValueMust(featureElemType, []attr.Value{})
	emptyBaseSet := types.SetValueMust(types.StringType, []attr.Value{})
	unknownFeatureSet := types.SetUnknown(featureElemType)
	unknownBaseSet := types.SetUnknown(types.StringType)

	tests := []struct {
		name       string
		kibanaElem attr.Value
		wantError  bool
	}{
		{
			name: "known non-empty feature",
			kibanaElem: types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
				attrSpaces:  spacesSet,
				attrBase:    emptyBaseSet,
				attrFeature: knownFeatureSet,
			}),
		},
		{
			name: "known non-empty base",
			kibanaElem: types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
				attrSpaces:  spacesSet,
				attrBase:    knownBaseSet,
				attrFeature: emptyFeatureSet,
			}),
		},
		{
			name: "known empty base and feature",
			kibanaElem: types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
				attrSpaces:  spacesSet,
				attrBase:    emptyBaseSet,
				attrFeature: emptyFeatureSet,
			}),
			wantError: true,
		},
		{
			name: "unknown feature set",
			kibanaElem: types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
				attrSpaces:  spacesSet,
				attrBase:    emptyBaseSet,
				attrFeature: unknownFeatureSet,
			}),
		},
		{
			name: "unknown base set",
			kibanaElem: types.ObjectValueMust(kibanaBlockAttrTypes(), map[string]attr.Value{
				attrSpaces:  spacesSet,
				attrBase:    unknownBaseSet,
				attrFeature: emptyFeatureSet,
			}),
		},
		{
			name:       "fully unknown kibana element",
			kibanaElem: types.ObjectUnknown(kibanaBlockAttrTypes()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			diags := validateResourceConfig(t, tt.kibanaElem)
			if tt.wantError {
				require.True(t, diags.HasError(), "expected privilege validation error")
				found := false
				for _, d := range diags.Errors() {
					if strings.Contains(d.Detail(), "Either one of the `feature` or `base` privileges must be set for kibana role!") {
						found = true
						break
					}
				}
				assert.True(t, found, "expected missing-privilege diagnostic")
				return
			}
			assert.False(t, diags.HasError(), "expected no validation errors, got: %v", diags)
		})
	}
}

func validateResourceConfig(t *testing.T, kibanaElem attr.Value) diag.Diagnostics {
	t.Helper()

	r := NewResource()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	kibanaSet := types.SetValueMust(kibanaBlockObjectType(), []attr.Value{kibanaElem})
	esObj := types.ObjectValueMust(elasticsearchResourceAttrTypes(), map[string]attr.Value{
		attrCluster:       types.SetNull(types.StringType),
		attrIndices:       types.SetNull(types.ObjectType{AttrTypes: esIndexResourceAttrTypes()}),
		attrRemoteIndices: types.SetNull(types.ObjectType{AttrTypes: esRemoteIndexResourceAttrTypes()}),
		attrRunAs:         types.SetNull(types.StringType),
	})

	attrTypes := schemaResp.Schema.Type().TerraformType(context.Background())
	kibanaTfVal, err := kibanaSet.ToTerraformValue(context.Background())
	require.NoError(t, err)
	esTfVal, err := esObj.ToTerraformValue(context.Background())
	require.NoError(t, err)

	objType := attrTypes.(tftypes.Object)
	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, typ := range objType.AttributeTypes {
		attrs[name] = tftypes.NewValue(typ, nil)
	}
	attrs["name"] = tftypes.NewValue(tftypes.String, "test-role")
	attrs["elasticsearch"] = esTfVal
	attrs["kibana"] = kibanaTfVal

	rawConfig := tftypes.NewValue(attrTypes, attrs)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    rawConfig,
	}

	resp := &resource.ValidateConfigResponse{}
	configValidator{}.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: config}, resp)
	return resp.Diagnostics
}
