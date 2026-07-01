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

package agentlesspolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// getSchema is a bare-minimum skeleton schema for Task 3 of the
// fleet-agentless-policy OpenSpec change. Only the identity attributes
// needed to back agentlessPolicyModel's entitycore.KibanaResourceModel
// implementation are defined here. The full schema (package, inputs,
// vars_json, cloud_connector, global_data_tags,
// additional_datastreams_permissions, var_group_selections,
// create_dataset_templates, force/force_delete, created_at/updated_at,
// experimental notice -- see specs/fleet-agentless-policy/spec.md, "Schema
// attributes") is added in Task 4.
func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages Fleet agentless policies. " +
			"Skeleton only pending full implementation; see openspec/changes/fleet-agentless-policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The composite ID of the agentless policy: `<space_id>/<policy_id>`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				MarkdownDescription: "The agentless policy (package policy) ID. Server-assigned if omitted; forces replacement on change.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_ids": schema.SetAttribute{
				Computed:            true,
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The list of spaces the agentless policy belongs to; forces replacement on change.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
