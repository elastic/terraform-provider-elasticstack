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

package entitycore

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// importTestModel is a minimal model with id and name for testing ImportStateWithNameAttribute.
type importTestModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

var importTestSchema = schema.Schema{
	Attributes: map[string]schema.Attribute{
		"id":   schema.StringAttribute{Computed: true},
		"name": schema.StringAttribute{Required: true},
	},
}

var importTestTFType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"id":   tftypes.String,
		"name": tftypes.String,
	},
}

func emptyImportTestState() tfsdk.State {
	return tfsdk.State{
		Raw: tftypes.NewValue(importTestTFType, map[string]tftypes.Value{
			"id":   tftypes.NewValue(tftypes.String, nil),
			"name": tftypes.NewValue(tftypes.String, nil),
		}),
		Schema: importTestSchema,
	}
}

func TestImportStateWithNameAttribute_compositeID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := resource.ImportStateRequest{ID: "cluster-uuid/my-pattern"}
	resp := &resource.ImportStateResponse{State: emptyImportTestState()}

	ImportStateWithNameAttribute(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var model importTestModel
	diags := resp.State.Get(ctx, &model)
	require.False(t, diags.HasError(), diags)

	assert.Equal(t, "cluster-uuid/my-pattern", model.ID.ValueString(), "composite ID should be stored in id")
	assert.Equal(t, "my-pattern", model.Name.ValueString(), "resource name should be extracted into name")
}

func TestImportStateWithNameAttribute_plainID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := resource.ImportStateRequest{ID: "my-pattern"}
	resp := &resource.ImportStateResponse{State: emptyImportTestState()}

	ImportStateWithNameAttribute(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), resp.Diagnostics)

	var model importTestModel
	diags := resp.State.Get(ctx, &model)
	require.False(t, diags.HasError(), diags)

	assert.Equal(t, "my-pattern", model.Name.ValueString(), "plain ID should be stored in name")
	assert.True(t, model.ID.IsNull(), "id should remain null for plain import")
}
