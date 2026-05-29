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

package cloudconnector

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func TestResource_ImportState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	r, ok := any(newResource()).(resource.ResourceWithImportState)
	require.True(t, ok)

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		resp := emptyImportStateResponse(ctx, t, r)

		r.ImportState(ctx, resource.ImportStateRequest{ID: "default/my-connector"}, resp)
		require.False(t, resp.Diagnostics.HasError())

		var id, spaceID, connectorID types.String
		resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("id"), &id)...)
		resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("space_id"), &spaceID)...)
		resp.Diagnostics.Append(resp.State.GetAttribute(ctx, path.Root("cloud_connector_id"), &connectorID)...)
		require.False(t, resp.Diagnostics.HasError())
		require.Equal(t, "default/my-connector", id.ValueString())
		require.Equal(t, "default", spaceID.ValueString())
		require.Equal(t, "my-connector", connectorID.ValueString())
	})

	t.Run("plain id without slash", func(t *testing.T) {
		t.Parallel()
		resp := emptyImportStateResponse(ctx, t, r)

		r.ImportState(ctx, resource.ImportStateRequest{ID: "plain-id"}, resp)
		require.True(t, resp.Diagnostics.HasError())
	})

	t.Run("empty space segment", func(t *testing.T) {
		t.Parallel()
		resp := emptyImportStateResponse(ctx, t, r)

		r.ImportState(ctx, resource.ImportStateRequest{ID: "/my-connector"}, resp)
		require.True(t, resp.Diagnostics.HasError())
		require.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Invalid import ID")
	})
}

func emptyImportStateResponse(ctx context.Context, t *testing.T, r resource.Resource) *resource.ImportStateResponse {
	t.Helper()

	var sr resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sr)
	require.False(t, sr.Diagnostics.HasError())

	raw, err := zeroTerraformValue(sr.Schema.Type().TerraformType(ctx))
	require.NoError(t, err)

	return &resource.ImportStateResponse{
		State: tfsdk.State{
			Schema: sr.Schema,
			Raw:    raw,
		},
	}
}

func zeroTerraformValue(typ tftypes.Type) (tftypes.Value, error) {
	if typ == nil {
		return tftypes.Value{}, fmt.Errorf("nil tftypes.Type")
	}

	switch t := typ.(type) {
	case tftypes.Object:
		m := make(map[string]tftypes.Value, len(t.AttributeTypes))
		for name, at := range t.AttributeTypes {
			v, err := zeroTerraformValue(at)
			if err != nil {
				return tftypes.Value{}, err
			}
			m[name] = v
		}
		return tftypes.NewValue(t, m), nil
	case tftypes.Tuple:
		els := make([]tftypes.Value, len(t.ElementTypes))
		for i, et := range t.ElementTypes {
			v, err := zeroTerraformValue(et)
			if err != nil {
				return tftypes.Value{}, err
			}
			els[i] = v
		}
		return tftypes.NewValue(t, els), nil
	case tftypes.List, tftypes.Set, tftypes.Map:
		return tftypes.NewValue(t, nil), nil
	default:
		if typ.Is(tftypes.String) || typ.Is(tftypes.Number) || typ.Is(tftypes.Bool) || typ.Is(tftypes.DynamicPseudoType) {
			return tftypes.NewValue(typ, nil), nil
		}
		return tftypes.Value{}, fmt.Errorf("unsupported tftypes.Type %v (%[1]T)", typ)
	}
}
