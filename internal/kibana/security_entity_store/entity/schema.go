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

package entity

import (
	"context"

	jsontypes "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a single entity record in the Kibana Security Entity Store.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:   "Computed resource identifier in the format <space_id>/<entity_id>.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"space_id": schema.StringAttribute{
				Description:   "An identifier for the Kibana space. If omitted, the default space is used.",
				Optional:      true,
				Computed:      true,
				Default:       stringdefault.StaticString(defaultSpaceID),
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"entity_type": schema.StringAttribute{
				Description:   "The type of entity. Must be one of: user, host, service, generic.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf("user", "host", "service", "generic"),
				},
			},
			"entity_id": schema.StringAttribute{
				Description:   "Unique identifier for this entity. Must match the entity.id field in the typed entity block or entity_json when supplied.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"timestamp": schema.StringAttribute{
				Description: "The time the entity record was last updated. Maps to @timestamp in the API body.",
				Optional:    true,
				Computed:    true,
			},
			"labels": schema.MapAttribute{
				Description: "Labels associated with the entity as a map of string to string.",
				Optional:    true,
				ElementType: types.StringType,
				Validators:  []validator.Map{},
			},
			"tags": schema.SetAttribute{
				Description: "Tags associated with the entity.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"document_json": schema.StringAttribute{
				Description: "Canonical JSON (sorted keys) containing the full entity document as read back from Kibana.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"response_json": schema.StringAttribute{
				Description: "Raw API response body serialized as normalized JSON for troubleshooting.",
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
			},
			"force": schema.BoolAttribute{
				Description: "When true, passes force=true on PUT updates.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// JSON fallbacks
			"entity_json": schema.StringAttribute{
				Description: "JSON fallback for the entity block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot(attrEntity)),
				},
			},
			"host_json": schema.StringAttribute{
				Description: "JSON fallback for the host block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("host")),
				},
			},
			"user_json": schema.StringAttribute{
				Description: "JSON fallback for the user block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("user")),
				},
			},
			"service_json": schema.StringAttribute{
				Description: "JSON fallback for the service block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("service")),
				},
			},
			"cloud_json": schema.StringAttribute{
				Description: "JSON fallback for the cloud block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("cloud")),
				},
			},
			"asset_json": schema.StringAttribute{
				Description: "JSON fallback for the asset block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot(attrAsset)),
				},
			},
			"orchestrator_json": schema.StringAttribute{
				Description: "JSON fallback for the orchestrator block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("orchestrator")),
				},
			},
			"event_json": schema.StringAttribute{
				Description: "JSON fallback for the event block.",
				Optional:    true,
				CustomType:  jsontypes.NormalizedType{},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("event")),
				},
			},
			"labels_json": schema.StringAttribute{
				Description: "JSON fallback for labels. Supports non-string values.",
				Optional:    true,
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("labels")),
				},
			},
			// Typed blocks
			attrEntity: schema.SingleNestedAttribute{
				Description: "Core entity fields shared across all entity types.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Description: "Unique identifier for this entity.",
						Required:    true,
					},
					attrName: schema.StringAttribute{
						Description: "Human-readable name of the entity.",
						Optional:    true,
						Computed:    true,
					},
					attrType: schema.StringAttribute{
						Description: "The entity type.",
						Optional:    true,
						Computed:    true,
					},
					"sub_type": schema.StringAttribute{
						Description: "Optional sub-type classification for the entity.",
						Optional:    true,
						Computed:    true,
					},
					"source": schema.SetAttribute{
						Description: "The sources that produced this entity record.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"attributes": schema.SingleNestedAttribute{
						Description: "Boolean flags describing characteristics of the entity.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrAsset: schema.BoolAttribute{
								Description: "Whether the entity is classified as an asset.",
								Optional:    true,
								Computed:    true,
							},
							"managed": schema.BoolAttribute{
								Description: "Whether the entity is managed (for example, via a directory service).",
								Optional:    true,
								Computed:    true,
							},
							"privileged": schema.BoolAttribute{
								Description: "Whether the entity has elevated privileges.",
								Optional:    true,
								Computed:    true,
							},
							"mfa_enabled": schema.BoolAttribute{
								Description: "Whether multi-factor authentication is enabled for the entity.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"behaviors": schema.SingleNestedAttribute{
						Description: "Boolean flags indicating observed behavioral signals.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"brute_force_victim": schema.BoolAttribute{
								Description: "Whether the entity has been targeted by brute-force attacks.",
								Optional:    true,
								Computed:    true,
							},
							"new_country_login": schema.BoolAttribute{
								Description: "Whether the entity has logged in from a new country.",
								Optional:    true,
								Computed:    true,
							},
							"used_usb_device": schema.BoolAttribute{
								Description: "Whether the entity has used a USB device.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"lifecycle": schema.SingleNestedAttribute{
						Description: "Timestamps tracking the entity lifecycle.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"first_seen": schema.StringAttribute{
								Description: "When the entity was first observed.",
								Optional:    true,
								Computed:    true,
							},
							"last_seen": schema.StringAttribute{
								Description: "When the entity was last observed.",
								Optional:    true,
								Computed:    true,
							},
							"last_activity": schema.StringAttribute{
								Description: "When the entity last generated activity.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					attrRisk: schema.SingleNestedAttribute{
						Description: "Risk scoring information for the entity.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrCalculatedLevel: schema.StringAttribute{
								Description: descCalculatedLevel,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScore: schema.Float64Attribute{
								Description: descCalculatedScore,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScoreNorm: schema.Float64Attribute{
								Description: descCalculatedScoreNorm,
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"relationships": schema.SingleNestedAttribute{
						Description: "Connections between this entity and other entities.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"owned_by": schema.SetAttribute{
								Description: "Entity IDs that own this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"owns": schema.SetAttribute{
								Description: "Entity IDs owned by this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"supervised_by": schema.SetAttribute{
								Description: "Entity IDs that supervise this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"supervises": schema.SetAttribute{
								Description: "Entity IDs supervised by this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"depends_on": schema.SetAttribute{
								Description: "Entity IDs this entity depends on.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"dependent_of": schema.SetAttribute{
								Description: "Entity IDs that depend on this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"communicates_with": schema.SetAttribute{
								Description: "Entity IDs this entity communicates with.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"accesses_frequently": schema.SetAttribute{
								Description: "Entity IDs this entity accesses frequently.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"accessed_frequently_by": schema.SetAttribute{
								Description: "Entity IDs that frequently access this entity.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
							"accesses_infrequently": schema.SetAttribute{
								Description: "Entity IDs this entity accesses infrequently.",
								Optional:    true,
								Computed:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("entity_json")),
				},
			},
			"host": schema.SingleNestedAttribute{
				Description: "ECS host fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					attrName: schema.StringAttribute{
						Description: "Primary host name.",
						Required:    true,
					},
					attrDomain: schema.SetAttribute{
						Description: "Observed host domains.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"hostname": schema.SetAttribute{
						Description: "Observed hostnames.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"id": schema.SetAttribute{
						Description: "Observed host IDs.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"ip": schema.SetAttribute{
						Description: "Observed IP addresses.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"mac": schema.SetAttribute{
						Description: "Observed MAC addresses.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					attrType: schema.SetAttribute{
						Description: "Observed host types.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"architecture": schema.SetAttribute{
						Description: "Observed CPU architectures.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"os": schema.SingleNestedAttribute{
						Description: "Elastic Common Schema (ECS) host.os fields collected on the entity.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"family": schema.StringAttribute{
								Description: "Operating system family.",
								Optional:    true,
								Computed:    true,
							},
							"full": schema.StringAttribute{
								Description: "Full operating system name.",
								Optional:    true,
								Computed:    true,
							},
							"kernel": schema.StringAttribute{
								Description: "Kernel version.",
								Optional:    true,
								Computed:    true,
							},
							attrName: schema.StringAttribute{
								Description: "Operating system name.",
								Optional:    true,
								Computed:    true,
							},
							"platform": schema.StringAttribute{
								Description: "Operating system platform.",
								Optional:    true,
								Computed:    true,
							},
							attrType: schema.StringAttribute{
								Description: "Operating system type.",
								Optional:    true,
								Computed:    true,
							},
							"version": schema.StringAttribute{
								Description: "Operating system version.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					attrRisk: schema.SingleNestedAttribute{
						Description: "Risk scoring information for the host.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrCalculatedLevel: schema.StringAttribute{
								Description: descCalculatedLevel,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScore: schema.Float64Attribute{
								Description: descCalculatedScore,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScoreNorm: schema.Float64Attribute{
								Description: descCalculatedScoreNorm,
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("host_json")),
				},
			},
			"user": schema.SingleNestedAttribute{
				Description: "ECS user fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					attrName: schema.StringAttribute{
						Description: "Primary user name.",
						Required:    true,
					},
					attrDomain: schema.SetAttribute{
						Description: "Observed user domains.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					attrEmail: schema.SetAttribute{
						Description: "Observed email addresses.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"full_name": schema.SetAttribute{
						Description: "Observed full names of the user.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"hash": schema.SetAttribute{
						Description: "Observed user hashes.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"id": schema.SetAttribute{
						Description: "Observed user IDs.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					"roles": schema.SetAttribute{
						Description: "Observed roles assigned to the user.",
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
					},
					attrRisk: schema.SingleNestedAttribute{
						Description: "Risk scoring information for the user.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrCalculatedLevel: schema.StringAttribute{
								Description: descCalculatedLevel,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScore: schema.Float64Attribute{
								Description: descCalculatedScore,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScoreNorm: schema.Float64Attribute{
								Description: descCalculatedScoreNorm,
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("user_json")),
				},
			},
			"service": schema.SingleNestedAttribute{
				Description: "ECS service fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					attrName: schema.StringAttribute{
						Description: "Primary service name.",
						Required:    true,
					},
					attrRisk: schema.SingleNestedAttribute{
						Description: "Risk scoring information for the service.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrCalculatedLevel: schema.StringAttribute{
								Description: descCalculatedLevel,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScore: schema.Float64Attribute{
								Description: descCalculatedScore,
								Optional:    true,
								Computed:    true,
							},
							attrCalculatedScoreNorm: schema.Float64Attribute{
								Description: descCalculatedScoreNorm,
								Optional:    true,
								Computed:    true,
							},
						},
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("service_json")),
				},
			},
			"orchestrator": schema.SingleNestedAttribute{
				Description: "Orchestrator fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					attrName: schema.StringAttribute{
						Description: "Orchestrator name.",
						Optional:    true,
						Computed:    true,
					},
					attrType: schema.StringAttribute{
						Description: "Orchestrator type.",
						Optional:    true,
						Computed:    true,
					},
					"namespace": schema.StringAttribute{
						Description: "Orchestrator namespace.",
						Optional:    true,
						Computed:    true,
					},
					"cluster_id": schema.StringAttribute{
						Description: "Cluster identifier.",
						Optional:    true,
						Computed:    true,
					},
					"cluster_name": schema.StringAttribute{
						Description: "Cluster name.",
						Optional:    true,
						Computed:    true,
					},
					"cluster_version": schema.StringAttribute{
						Description: "Cluster version.",
						Optional:    true,
						Computed:    true,
					},
					"resource_id": schema.StringAttribute{
						Description: "Resource identifier.",
						Optional:    true,
						Computed:    true,
					},
					"resource_name": schema.StringAttribute{
						Description: "Resource name.",
						Optional:    true,
						Computed:    true,
					},
					"resource_type": schema.StringAttribute{
						Description: "Resource type.",
						Optional:    true,
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("orchestrator_json")),
				},
			},
			"cloud": schema.SingleNestedAttribute{
				Description: "Cloud fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					attrProvider: schema.StringAttribute{
						Description: "Cloud provider.",
						Optional:    true,
						Computed:    true,
					},
					"region": schema.StringAttribute{
						Description: "Cloud region.",
						Optional:    true,
						Computed:    true,
					},
					"account_id": schema.StringAttribute{
						Description: "Cloud account identifier.",
						Optional:    true,
						Computed:    true,
					},
					"account_name": schema.StringAttribute{
						Description: "Cloud account name.",
						Optional:    true,
						Computed:    true,
					},
					"project_id": schema.StringAttribute{
						Description: "Cloud project identifier.",
						Optional:    true,
						Computed:    true,
					},
					"project_name": schema.StringAttribute{
						Description: "Cloud project name.",
						Optional:    true,
						Computed:    true,
					},
					"service_name": schema.StringAttribute{
						Description: "Cloud service name.",
						Optional:    true,
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("cloud_json")),
				},
			},
			"event": schema.SingleNestedAttribute{
				Description: "Event fields collected on the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"category": schema.StringAttribute{
						Description: "Event category.",
						Optional:    true,
						Computed:    true,
					},
					attrType: schema.StringAttribute{
						Description: "Event type.",
						Optional:    true,
						Computed:    true,
					},
					"dataset": schema.StringAttribute{
						Description: "Event dataset.",
						Optional:    true,
						Computed:    true,
					},
					"kind": schema.StringAttribute{
						Description: "Event kind.",
						Optional:    true,
						Computed:    true,
					},
					"outcome": schema.StringAttribute{
						Description: "Event outcome.",
						Optional:    true,
						Computed:    true,
					},
					attrProvider: schema.StringAttribute{
						Description: "Event provider.",
						Optional:    true,
						Computed:    true,
					},
					"action": schema.StringAttribute{
						Description: "Event action.",
						Optional:    true,
						Computed:    true,
					},
					"code": schema.StringAttribute{
						Description: "Event code.",
						Optional:    true,
						Computed:    true,
					},
					"reference": schema.StringAttribute{
						Description: "Event reference.",
						Optional:    true,
						Computed:    true,
					},
					attrReason: schema.StringAttribute{
						Description: "Event reason.",
						Optional:    true,
						Computed:    true,
					},
					"severity": schema.StringAttribute{
						Description: "Event severity.",
						Optional:    true,
						Computed:    true,
					},
					"timezone": schema.StringAttribute{
						Description: "Event timezone.",
						Optional:    true,
						Computed:    true,
					},
					"url": schema.StringAttribute{
						Description: "Event URL.",
						Optional:    true,
						Computed:    true,
					},
					"ingested": schema.StringAttribute{
						Description: "When the event was ingested into Elasticsearch.",
						Optional:    true,
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("event_json")),
				},
			},
			attrAsset: schema.SingleNestedAttribute{
				Description: "Asset metadata associated with the entity.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"criticality": schema.StringAttribute{
						Description: "Asset criticality level.",
						Optional:    true,
						Computed:    true,
					},
					"criticality_feedback": schema.SingleNestedAttribute{
						Description: "Feedback on the asset criticality.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"notes": schema.StringAttribute{
								Description: "Feedback notes.",
								Optional:    true,
								Computed:    true,
							},
							attrReason: schema.StringAttribute{
								Description: "Feedback reason.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					"owner": schema.SingleNestedAttribute{
						Description: "Asset owner information.",
						Optional:    true,
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							attrName: schema.StringAttribute{
								Description: "Owner name.",
								Optional:    true,
								Computed:    true,
							},
							"department": schema.StringAttribute{
								Description: "Owner department.",
								Optional:    true,
								Computed:    true,
							},
							attrEmail: schema.StringAttribute{
								Description: "Owner email.",
								Optional:    true,
								Computed:    true,
							},
							"ext": schema.StringAttribute{
								Description: "Owner extension.",
								Optional:    true,
								Computed:    true,
							},
						},
					},
					attrValue: schema.Float64Attribute{
						Description: "Asset value.",
						Optional:    true,
						Computed:    true,
					},
				},
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.MatchRoot("asset_json")),
				},
			},
		},
	}
}
