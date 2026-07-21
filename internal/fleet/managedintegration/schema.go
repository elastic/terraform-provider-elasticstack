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

package managedintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// attrName is the schema attribute key "name", reused across the top-level
// identity attribute, the package object, and the cloud_connector object.
const attrName = "name"

// getSchema defines the elasticstack_fleet_managed_integration resource schema
// (openspec/changes/fleet-managed-integration/specs/fleet-managed-integration/
// spec.md, "Schema attributes"). CRUD population lives in models_convert.go
// and create/read/update/delete. Version gating is in models.go; the
// deployment-topology preflight is in create.go/topology.go.
//
// kibana_connection: unlike internal/fleet/integration_policy
// and internal/fleet/agentpolicy (which implement resource.Resource's Schema
// method directly and so must add the `kibana_connection` block themselves),
// this resource is built on the entitycore.KibanaResource[T] envelope (see
// resource.go), which injects the `kibana_connection` block (and the
// `timeouts` attribute) into whatever schema.Schema this factory returns --
// see baseResourceEnvelope.Schema in internal/entitycore/base_envelope.go.
// The canonical pattern to mirror here is therefore
// internal/fleet/proxy/schema.go (also envelope-based), which likewise
// defines no Blocks at all.
func getSchema(_ context.Context) schema.Schema {
	varsAreSensitive := debugutils.IsSensitiveInSchema()
	return schema.Schema{
		MarkdownDescription: "Manages Fleet managed integrations, which provision agent runtime capacity in Elastic's " +
			"own cloud infrastructure instead of on a host running Elastic Agent. " +
			"**This resource is experimental**: the underlying Fleet managed integrations API requires Kibana " +
			"9.5.0 and its behavior may change in future Kibana releases. " +
			"It is only supported on **Elastic Cloud Hosted** and **Serverless** (Security or Observability) " +
			"deployments; self-managed (on-premises) Kibana is not supported, and this resource refuses to run " +
			"against a self-managed deployment it can positively identify as such.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the managed integration: `<space_id>/<policy_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The managed integration ID. Server-assigned if omitted; forces replacement on change.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrName: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the managed integration; updatable in-place.",
			},
			"description": schema.StringAttribute{
				Optional: true,
				MarkdownDescription: "The description of the managed integration; updatable in-place. " +
					"An explicit empty string is rejected: it is indistinguishable from \"unset\" once " +
					"round-tripped through the API (Kibana returns an omitted/empty description as `\"\"`, " +
					"which this provider folds back to null), so setting `description = \"\"` would otherwise " +
					"produce a permanent, non-converging diff. Omit the attribute instead of setting it to `\"\"`.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"namespace": schema.StringAttribute{
				Computed: true,
				Optional: true,
				MarkdownDescription: "The namespace of the managed integration; forces replacement on change. " +
					"An explicit empty string is rejected for the same reason as `description`: it is " +
					"indistinguishable from \"unset\" once round-tripped through the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"space_ids": schema.SetAttribute{
				Computed:            true,
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The list of spaces the managed integration belongs to; defaults to `[\"default\"]`; forces replacement on change.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"package": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The Fleet integration package this managed integration is based on.",
				Attributes: map[string]schema.Attribute{
					attrName: schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The package name; forces replacement on change.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"version": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "The package version; updatable in-place.",
					},
					"title": schema.StringAttribute{
						Computed: true,
						Optional: true,
						MarkdownDescription: "The package title. If omitted, Kibana populates it from the package registry. " +
							"Updatable in-place (not `RequiresReplace`).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"policy_template": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The policy template within the package to use; forces replacement on change.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vars_json": schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				Sensitive:  varsAreSensitive,
				CustomType: policyshape.NewVarsJSONType(lookupCachedPackageInfo),
				MarkdownDescription: customtypes.DescriptionWithContextWarning(
					"Integration-level variables as JSON. Variables vary depending on the integration package. Updatable in-place.",
				),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"var_group_selections": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				MarkdownDescription: "Top-level variable group selections, mapping group name to selected option; updatable in-place. " +
					"Modeled at the top level only in v1; per-stream var_group_selections is deferred to a follow-up change.",
			},
			"inputs": schema.MapNestedAttribute{
				Computed:            true,
				Optional:            true,
				CustomType:          policyshape.NewInputsType(agentlessInputType()),
				MarkdownDescription: "Policy inputs mapped by input type ID; updatable in-place.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: getInputsNestedObject(varsAreSensitive),
			},
			"cloud_connector": schema.SingleNestedAttribute{
				Optional: true,
				MarkdownDescription: "References an existing cloud connector for cross-account access. " +
					"Changing any field forces replacement of the entire `cloud_connector` block.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Whether the cloud connector is enabled for this policy.",
					},
					"cloud_connector_id": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The ID of an existing cloud connector to associate with this policy.",
					},
					attrName: schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The name of the cloud connector.",
					},
					"target_csp": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The target cloud service provider for the cloud connector. One of `aws`, `azure`, or `gcp`.",
						Validators: []validator.String{
							stringvalidator.OneOf("aws", "azure", "gcp"),
						},
					},
				},
			},
			"global_data_tags": schema.MapNestedAttribute{
				Optional: true,
				MarkdownDescription: "Global data tags applied to the managed integration's data streams; updatable in-place. " +
					"Keyed by tag name; set exactly one of `string_value` or `number_value` per entry.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						globalDataTagStringValueAttr: schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "String value for the tag. If this is set, `number_value` must not be defined.",
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName(globalDataTagNumberValueAttr)),
								stringvalidator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(globalDataTagStringValueAttr),
									path.MatchRelative().AtParent().AtName(globalDataTagNumberValueAttr),
								),
							},
						},
						globalDataTagNumberValueAttr: schema.Float32Attribute{
							Optional:            true,
							MarkdownDescription: "Number value for the tag. If this is set, `string_value` must not be defined.",
							Validators: []validator.Float32{
								float32validator.ConflictsWith(path.MatchRelative().AtParent().AtName(globalDataTagStringValueAttr)),
								float32validator.AtLeastOneOf(
									path.MatchRelative().AtParent().AtName(globalDataTagStringValueAttr),
									path.MatchRelative().AtParent().AtName(globalDataTagNumberValueAttr),
								),
							},
						},
					},
				},
			},
			"additional_datastreams_permissions": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Additional data stream permissions to grant beyond the package's defaults; updatable in-place.",
			},
			"create_dataset_templates": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Whether to create dataset templates when creating the policy. Create-only: sent on the create " +
					"request only, not read back from the API. Changes after creation are a no-op until the resource is recreated.",
			},
			"force": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Force the create operation. Create-only: sent on the create request only " +
					"and not read back from the API.",
			},
			"force_delete": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "Force deletion of the policy, passed as `?force=true` on the delete request. " +
					"Defaults to `false`.",
			},
			"skip_topology_check": schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "Skips the deployment-topology preflight check. Use only if you are certain " +
					"this is running against a supported Elastic Cloud Hosted or Serverless deployment and the " +
					"automatic detection is producing a false positive (e.g. due to non-standard network routing " +
					"such as PrivateLink). Does not weaken version gating (Kibana 9.5.0+ is still enforced) -- it " +
					"only bypasses the topology heuristic. Defaults to `false`. Create-only: consulted only during " +
					"Create and not read back from the API.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation timestamp of the managed integration (ISO 8601).",
				// UseStateForUnknown is correct (and safe in every Update
				// scenario, real or short-circuited) because created_at never
				// changes after the resource is created -- Kibana never
				// updates it. See update.go's onlyCreateOnlyFlagsChanged doc
				// comment for why this attribute is nonetheless excluded from
				// that function's comparison chain: it's a belt-and-suspenders
				// fix, not a substitute for it.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The last-updated timestamp of the managed integration (ISO 8601).",
				// Deliberately NOT UseStateForUnknown, unlike created_at
				// above: updated_at legitimately changes on every real
				// Update (Kibana bumps it), so pre-committing the plan to
				// "this will stay equal to the prior state's value" is
				// actively wrong whenever a real update happens -- doing so
				// was verified empirically (via the acceptance test step in
				// acc_test.go) to produce a live "Provider produced
				// inconsistent result after apply: unexpected new value:
				// .updated_at" error on Kibana returning a genuinely new
				// updated_at after a real content change. Leaving this
				// Computed-only means it is Unknown ("known after apply") in
				// every Update plan -- slightly noisier, but always
				// consistent with whatever Update's read-after-write refresh
				// actually returns. See update.go's onlyCreateOnlyFlagsChanged
				// doc comment for how the create-only-flags short-circuit is
				// made correct without relying on this attribute ever being
				// Known in the plan.
			},
		},
	}
}

// agentlessInputType returns the policyshape.InputType used as the element
// type of the top-level `inputs` map. It reuses the shared package's
// InputType/StreamType wrapper types and AttrXxx attribute-name constants,
// but (unlike internal/fleet/integration_policy, which also surfaces a
// package-computed `defaults` object) it deliberately omits `defaults`: the
// spec for this resource (specs/fleet-agentless-policy/spec.md, "Schema
// attributes") does not model package-defaults introspection for inputs, so
// this uses a smaller attribute-types map made up entirely of shared
// building blocks rather than policyshape.InputElementType()'s
// defaults-inclusive map.
func agentlessInputType() policyshape.InputType {
	return policyshape.NewInputType(agentlessInputAttributeTypes())
}

func agentlessInputAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		policyshape.AttrEnabled:   types.BoolType,
		policyshape.AttrCondition: types.StringType,
		policyshape.AttrVars:      jsontypes.NormalizedType{},
		policyshape.AttrStreams: types.MapType{
			ElemType: policyshape.StreamType(),
		},
	}
}

func getInputsNestedObject(varsAreSensitive bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		CustomType: agentlessInputType(),
		Attributes: map[string]schema.Attribute{
			policyshape.AttrEnabled: schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable the input.",
			},
			policyshape.AttrCondition: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Agent condition expression to evaluate whether to apply this input.",
			},
			policyshape.AttrVars: schema.StringAttribute{
				Computed:   true,
				Optional:   true,
				CustomType: jsontypes.NormalizedType{},
				Sensitive:  varsAreSensitive,
				MarkdownDescription: "Input-level variables as JSON. Computed (not purely Optional): some packages " +
					"(e.g. cloud_security_posture/CSPM) populate informational input-level vars " +
					"(such as CloudFormation quick-create template URLs) that are always present in the API " +
					"response regardless of configuration; Computed with UseStateForUnknown lets those flow " +
					"through without requiring the user to declare them.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			policyshape.AttrStreams: schema.MapNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Input streams mapped by stream ID.",
				NestedObject:        getInputStreamNestedObject(varsAreSensitive),
			},
		},
	}
}

func getInputStreamNestedObject(varsAreSensitive bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			policyshape.AttrEnabled: schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable the stream.",
			},
			policyshape.AttrCondition: schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Agent condition expression to evaluate whether to apply this stream.",
			},
			policyshape.AttrVars: schema.StringAttribute{
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
				Sensitive:           varsAreSensitive,
				MarkdownDescription: "Stream-level variables as JSON.",
			},
		},
	}
}
