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

package integration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationModelV0 struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Version                   types.String `tfsdk:"version"`
	Force                     types.Bool   `tfsdk:"force"`
	Prerelease                types.Bool   `tfsdk:"prerelease"`
	IgnoreMappingUpdateErrors types.Bool   `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool   `tfsdk:"skip_data_stream_rollover"`
	IgnoreConstraints         types.Bool   `tfsdk:"ignore_constraints"`
	SkipDestroy               types.Bool   `tfsdk:"skip_destroy"`
	SpaceIDs                  types.Set    `tfsdk:"space_ids"`
}

func getSchemaV0() *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required: true,
			},
			"force": schema.BoolAttribute{
				Optional: true,
			},
			"prerelease": schema.BoolAttribute{
				Optional: true,
			},
			"ignore_mapping_update_errors": schema.BoolAttribute{
				Optional: true,
			},
			"skip_data_stream_rollover": schema.BoolAttribute{
				Optional: true,
			},
			"ignore_constraints": schema.BoolAttribute{
				Optional: true,
			},
			"skip_destroy": schema.BoolAttribute{
				Optional: true,
			},
			"space_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *integrationResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: getSchemaV0(),
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var priorState integrationModelV0

				resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgradedState := integrationModel{
					ID:                        priorState.ID,
					Name:                      priorState.Name,
					Version:                   priorState.Version,
					Force:                     priorState.Force,
					Prerelease:                priorState.Prerelease,
					IgnoreMappingUpdateErrors: priorState.IgnoreMappingUpdateErrors,
					SkipDataStreamRollover:    priorState.SkipDataStreamRollover,
					IgnoreConstraints:         priorState.IgnoreConstraints,
					SkipDestroy:               priorState.SkipDestroy,
					SpaceID:                   types.StringNull(),
				}

				if !priorState.SpaceIDs.IsNull() && !priorState.SpaceIDs.IsUnknown() {
					var spaceIDs []string
					resp.Diagnostics.Append(priorState.SpaceIDs.ElementsAs(ctx, &spaceIDs, false)...)
					if resp.Diagnostics.HasError() {
						return
					}

					if len(spaceIDs) > 0 {
						selectedSpaceID := spaceIDs[0]
						ignoredSpaces := spaceIDs[1:]
						upgradedState.SpaceID = types.StringValue(spaceIDs[0])

						resp.Diagnostics.AddAttributeWarning(
							path.Root("space_ids"),
							"Multiple Space IDs Found",
							fmt.Sprintf(
								`The previous version of the resource allowed multiple space IDs to be set, although all but the first was ignored.
								The [%s] space from the previous state has been selected for the new "space_id" attribute, with %v spaces ignored. 
								If this is not correct, please update the resource configuration accordingly.`,
								selectedSpaceID, ignoredSpaces,
							),
						)
					}
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, upgradedState)...)
			},
		},
	}
}
