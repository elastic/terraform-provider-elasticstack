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

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Fleet cloud connectors. " +
			"See the [Fleet Cloud Connectors API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-fleet-cloud-connectors) for more information.",
		Attributes: map[string]schema.Attribute{
			attrID: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the cloud connector: `<space_id>/<cloud_connector_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrCloudConnectorID: schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The cloud connector ID. Server-assigned when omitted from configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrSpaceID: schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString("default"),
				MarkdownDescription: "An identifier for the space. If not provided, the default space is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			attrName: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The cloud connector name.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			attrCloudProvider: schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The cloud provider for this connector. One of `aws`, `azure`, or `gcp`. Changing this value forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(cloudProviderAWS, cloudProviderAzure, "gcp"),
				},
			},
			attrAccountType: schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The account type. One of `single-account` or `organization-account`. Updatable in place.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("single-account", "organization-account"),
				},
			},
			attrForceDelete: schema.BoolAttribute{
				Optional: true,
				MarkdownDescription: "When true, force-delete the connector even if it is still referenced by package policies. " +
					"The default of `false` returns an error from the API that includes the current `package_policy_count`.",
				Default: booldefault.StaticBool(false),
			},
			attrAWSBlock: schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Typed AWS authentication settings. Compiles to the same wire `vars` payload as the generic `vars` map. " +
					"Populated in state after Read when all modelled AWS keys are present and `cloud_provider` is `aws`.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					attrAWSRoleArn: schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The IAM role ARN Elastic assumes in the target AWS account.",
					},
					attrAWSExternalID: schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
						MarkdownDescription: "The external ID for the IAM trust relationship. Write-only: the value is sent to the API once and is never stored in state. " +
							"Plan-time drift detection uses a bcrypt hash in private state (see resource documentation).",
					},
					attrAWSExternalIDSecretRef: schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "The saved secret reference for `external_id` returned by the API after the secret is stored.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							attrSecretRefID: schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "The secret reference identifier.",
							},
							attrSecretRefIsSecretRef: schema.BoolAttribute{
								Computed:            true,
								MarkdownDescription: "Whether the value is a saved secret reference.",
							},
						},
					},
				},
			},
			attrAzureBlock: schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Typed Azure authentication settings. Compiles to the same wire `vars` payload as the generic `vars` map. " +
					"Populated in state after Read when all modelled Azure keys are present and `cloud_provider` is `azure`.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					attrAzureTenantID: schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Azure AD tenant ID.",
					},
					attrAzureClientID: schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Azure application (client) ID.",
					},
					attrAzureCloudConnectorID: schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The Azure cloud connector identifier stored in Fleet vars.",
					},
				},
			},
			attrVarsMap: schema.MapNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "Generic cloud connector variables keyed by integration package field name. Each element represents one API union arm. " +
					"Use this for GCP or when the typed blocks do not model every key returned by the API.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						varsElementValidator{},
					},
					Attributes: map[string]schema.Attribute{
						attrVarsString: schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Bare string var value (API union arm 1).",
						},
						attrVarsNumber: schema.Float64Attribute{
							Optional:            true,
							MarkdownDescription: "Bare numeric var value (API union arm 2).",
						},
						attrVarsBool: schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Bare boolean var value (API union arm 3).",
						},
						attrVarsType: schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Structured var type (API union arm 4), for example `text` or `password`.",
						},
						attrVarsFrozen: schema.BoolAttribute{
							Optional:            true,
							MarkdownDescription: "Whether the structured var is frozen. Valid only alongside `type`.",
						},
						attrVarsValue: schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "Plain string value for a structured var (API union arm 4).",
						},
						attrVarsSecretValue: schema.StringAttribute{
							Optional:            true,
							Sensitive:           true,
							WriteOnly:           true,
							MarkdownDescription: "Write-only secret value for a structured var (API union arm 4). The raw value is sent to the API once and is never stored in state.",
						},
						attrVarsSecretRef: schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Saved secret reference returned by the API for structured password vars.",
							PlanModifiers: []planmodifier.Object{
								objectplanmodifier.UseStateForUnknown(),
							},
							Attributes: map[string]schema.Attribute{
								attrSecretRefID: schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "The secret reference identifier.",
								},
								attrSecretRefIsSecretRef: schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Whether the value is a saved secret reference.",
								},
							},
						},
					},
				},
			},
			attrNamespace: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Fleet namespace associated with the connector.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrPackagePolicyCount: schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of package policies referencing this connector.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			attrVerificationStatus: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The connector verification status. May be null on first read because verification is asynchronous.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrVerificationStartedAt: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When connector verification started. May be null on first read because verification is asynchronous.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrVerificationFailedAt: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When connector verification failed, if applicable. May be null on first read because verification is asynchronous.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrCreatedAt: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the connector was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			attrUpdatedAt: schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the connector was last updated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
