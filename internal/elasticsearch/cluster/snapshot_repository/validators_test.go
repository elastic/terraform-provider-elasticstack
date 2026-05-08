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

package snapshot_repository

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func attrTypesToTfTypes(m map[string]attr.Type) map[string]tftypes.Type {
	out := make(map[string]tftypes.Type, len(m))
	for k, v := range m {
		out[k] = v.TerraformType(context.Background())
	}
	return out
}

func TestS3EndpointValidator(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"null", "", false},
		{"valid http", "http://s3.example.com", false},
		{"valid https", "https://s3.example.com:9000", false},
		{"invalid scheme", "ftp://s3.example.com", true},
		{"no scheme", "s3.example.com", true},
		{"no host", "http://", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := s3EndpointValidator{}
			resp := &validator.StringResponse{}
			v.ValidateString(context.Background(), validator.StringRequest{
				Path:        path.Root("endpoint"),
				ConfigValue: types.StringValue(tc.value),
			}, resp)
			if tc.wantError {
				require.True(t, resp.Diagnostics.HasError())
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfigExactlyOneType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	newSnapshotRepositoryResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	buildConfig := func(fsSet, urlSet, gcsSet, azureSet, s3Set, hdfsSet bool) tfsdk.Config {
		objVal := func(set bool, attrTypes map[string]attr.Type) tftypes.Value {
			tfAttrTypes := make(map[string]tftypes.Type, len(attrTypes))
			for k, v := range attrTypes {
				tfAttrTypes[k] = v.TerraformType(ctx)
			}
			objType := tftypes.Object{AttributeTypes: tfAttrTypes}
			if !set {
				return tftypes.NewValue(objType, nil)
			}
			vals := make(map[string]tftypes.Value)
			for k, ty := range tfAttrTypes {
				switch {
				case ty.Is(tftypes.Bool):
					vals[k] = tftypes.NewValue(ty, true)
				case ty.Is(tftypes.String):
					vals[k] = tftypes.NewValue(ty, "x")
				case ty.Is(tftypes.Number):
					vals[k] = tftypes.NewValue(ty, float64(1))
				default:
					vals[k] = tftypes.NewValue(ty, nil)
				}
			}
			return tftypes.NewValue(objType, vals)
		}

		schemaType := schemaResp.Schema.Type().TerraformType(ctx)
		allNull := map[string]tftypes.Value{}
		for attrName, attrType := range schemaType.(tftypes.Object).AttributeTypes {
			allNull[attrName] = tftypes.NewValue(attrType, nil)
		}

		allNull["name"] = tftypes.NewValue(tftypes.String, "test")
		allNull["verify"] = tftypes.NewValue(tftypes.Bool, true)
		allNull["fs"] = objVal(fsSet, fsAttrTypes())
		allNull["url"] = objVal(urlSet, urlAttrTypes())
		allNull["gcs"] = objVal(gcsSet, gcsAttrTypes())
		allNull["azure"] = objVal(azureSet, azureAttrTypes())
		allNull["s3"] = objVal(s3Set, s3AttrTypes())
		allNull["hdfs"] = objVal(hdfsSet, hdfsAttrTypes())

		raw := tftypes.NewValue(schemaType, allNull)
		return tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}
	}

	cases := []struct {
		name      string
		fs        bool
		url       bool
		gcs       bool
		azure     bool
		s3        bool
		hdfs      bool
		wantError bool
	}{
		{"exactly one fs", true, false, false, false, false, false, false},
		{"exactly one url", false, true, false, false, false, false, false},
		{"none set", false, false, false, false, false, false, true},
		{"multiple set", true, true, false, false, false, false, true},
		{"three set", false, false, true, true, true, false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v := validateConfigExactlyOneType{}
			req := resource.ValidateConfigRequest{Config: buildConfig(tc.fs, tc.url, tc.gcs, tc.azure, tc.s3, tc.hdfs)}
			resp := &resource.ValidateConfigResponse{}
			v.ValidateResource(ctx, req, resp)
			if tc.wantError {
				require.True(t, resp.Diagnostics.HasError(), "expected error")
			} else {
				require.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestValidateConfigExactlyOneType_UnknownSkips(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	newSnapshotRepositoryResource().Schema(ctx, resource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	schemaType := schemaResp.Schema.Type().TerraformType(ctx)
	allNull := map[string]tftypes.Value{}
	for attrName, attrType := range schemaType.(tftypes.Object).AttributeTypes {
		allNull[attrName] = tftypes.NewValue(attrType, nil)
	}
	allNull["name"] = tftypes.NewValue(tftypes.String, "test")
	allNull["verify"] = tftypes.NewValue(tftypes.Bool, true)

	fsTfTypes := attrTypesToTfTypes(fsAttrTypes())
	allNull["fs"] = tftypes.NewValue(tftypes.Object{AttributeTypes: fsTfTypes}, tftypes.UnknownValue)
	allNull["url"] = tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesToTfTypes(urlAttrTypes())}, nil)
	allNull["gcs"] = tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesToTfTypes(gcsAttrTypes())}, nil)
	allNull["azure"] = tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesToTfTypes(azureAttrTypes())}, nil)
	allNull["s3"] = tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesToTfTypes(s3AttrTypes())}, nil)
	allNull["hdfs"] = tftypes.NewValue(tftypes.Object{AttributeTypes: attrTypesToTfTypes(hdfsAttrTypes())}, nil)

	raw := tftypes.NewValue(schemaType, allNull)
	req := resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: raw}}
	resp := &resource.ValidateConfigResponse{}
	validateConfigExactlyOneType{}.ValidateResource(ctx, req, resp)
	require.False(t, resp.Diagnostics.HasError(), "validation should skip when a type block is unknown")
}
