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

package integrationpolicy

import (
	"context"
	_ "embed"

	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

//go:embed resource-description.md
var integrationPolicyDescription string

func (r *integrationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchemaV3()
}

func getSchemaV3() schema.Schema {
	varsAreSensitive := debugutils.IsSensitiveInSchema()
	return schema.Schema{
		Version:     3,
		Description: integrationPolicyDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrPolicyID: schema.StringAttribute{
				Description: "Unique identifier of the integration policy.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrName: schema.StringAttribute{
				Description: "The name of the integration policy.",
				Required:    true,
			},
			attrNamespace: schema.StringAttribute{
				Description: "The namespace of the integration policy.",
				Required:    true,
			},
			attrAgentPolicyID: schema.StringAttribute{
				Description: "ID of the agent policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Root(attrAgentPolicyIDs).Expression()),
				},
			},
			attrAgentPolicyIDs: schema.ListAttribute{
				Description: "List of agent policy IDs.",
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.Root(attrAgentPolicyID).Expression()),
					listvalidator.SizeAtLeast(1),
				},
			},
			attrDescription: schema.StringAttribute{
				Description: "The description of the integration policy.",
				Optional:    true,
			},
			attrForce: schema.BoolAttribute{
				Description: "Force operations, such as creation and deletion, to occur.",
				Optional:    true,
			},
			attrIntegrationName: schema.StringAttribute{
				Description: "The name of the integration package.",
				Required:    true,
			},
			attrIntegrationVersion: schema.StringAttribute{
				Description: "The version of the integration package.",
				Required:    true,
			},
			attrOutputID: schema.StringAttribute{
				Description: "The ID of the output to send data to. When not specified, the default output of the agent policy will be used.",
				Optional:    true,
			},
			attrVarsJSON: schema.StringAttribute{
				Description: customtypes.DescriptionWithContextWarning("Integration-level variables as JSON. Variables vary depending on the integration package."),
				CustomType:  policyshape.NewVarsJSONType(lookupCachedPackageInfo),
				Computed:    true,
				Optional:    true,
				Sensitive:   varsAreSensitive,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrSpaceIDs: schema.SetAttribute{
				Description: spaceIDsDescription,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"inputs": schema.MapNestedAttribute{
				Description: "Integration inputs mapped by input ID.",
				CustomType:  NewInputsType(NewInputType(getInputsAttributeTypes())),
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: getInputsNestedObject(varsAreSensitive),
			},
		},

		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		}}
}

func getInputsNestedObject(varsAreSensitive bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		CustomType: NewInputType(getInputsAttributeTypes()),
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				Description: "Enable the input.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			attrCondition: schema.StringAttribute{
				Description: "Agent condition expression to evaluate whether to apply this input.",
				Optional:    true,
			},
			attrVars: schema.StringAttribute{
				Description: "Input-level variables as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Sensitive:   varsAreSensitive,
			},
			attrDefaults: schema.SingleNestedAttribute{
				Description: "Input defaults.",
				Computed:    true,
				Default: objectdefault.StaticValue(basetypes.NewObjectNull(
					getInputDefaultsAttrTypes(),
				)),
				Attributes: map[string]schema.Attribute{
					attrVars: schema.StringAttribute{
						Description: "Input-level variable defaults as JSON.",
						CustomType:  jsontypes.NormalizedType{},
						Computed:    true,
					},
					attrStreams: schema.MapNestedAttribute{
						Description: "Stream-level defaults mapped by stream ID.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								attrEnabled: schema.BoolAttribute{
									Description: "Default enabled state for the stream.",
									Computed:    true,
								},
								attrVars: schema.StringAttribute{
									Description: "Stream-level variable defaults as JSON.",
									CustomType:  jsontypes.NormalizedType{},
									Computed:    true,
								},
							},
						},
					},
				},
			},
			attrStreams: schema.MapNestedAttribute{
				Description:  "Input streams mapped by stream ID.",
				Optional:     true,
				NestedObject: getInputStreamNestedObject(varsAreSensitive),
			},
		},
	}
}

func getInputStreamNestedObject(varsAreSensitive bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			attrEnabled: schema.BoolAttribute{
				Description: "Enable the stream.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			attrCondition: schema.StringAttribute{
				Description: "Agent condition expression to evaluate whether to apply this stream.",
				Optional:    true,
			},
			attrVars: schema.StringAttribute{
				Description: "Stream-level variables as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Sensitive:   varsAreSensitive,
			},
		},
	}
}

// getInputsElementType, getInputsAttributeTypes, getInputStreamType, and
// getInputDefaultsAttrTypes delegate to the shared policyshape package,
// which owns the canonical inputs/streams/defaults attribute-type structure
// (see internal/fleet/policyshape/attribute_types.go). They are kept here
// under their original names so schema.go, models.go, and this package's
// tests don't need a mechanical rename.

func getInputsElementType() InputType {
	return policyshape.InputElementType()
}

func getInputsAttributeTypes() map[string]attr.Type {
	return policyshape.InputAttributeTypes()
}

func getInputStreamType() attr.Type {
	return policyshape.StreamType()
}

func getInputDefaultsAttrTypes() map[string]attr.Type {
	return policyshape.InputDefaultsAttributeTypes()
}
