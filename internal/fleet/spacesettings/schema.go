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

package spacesettings

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const maxAllowedNamespacePrefixes = 10

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages the Fleet settings of a Kibana space, currently the namespace prefixes that integration policies in the space may use. " +
			"Requires Elastic Stack 9.1.0 or newer.\n\n" +
			"Fleet keeps exactly one settings object per space, so declare exactly one `elasticstack_fleet_space_settings` resource per space. " +
			"Multiple resources targeting the same space overwrite each other. " +
			"The resource neither creates the space nor checks that it exists, and Kibana accepts settings for a space that does not exist, so double-check `space_id`. " +
			"Destroying the resource resets `allowed_namespace_prefixes` to an empty set, which lifts the namespace restriction for the space.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource. Equal to `space_id`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "The ID of the space whose Fleet settings are managed. Changing it forces a new resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allowed_namespace_prefixes": schema.SetAttribute{
				Description: "Namespace prefixes that integration policies in the space are restricted to. " +
					"At most 10 unique prefixes; an empty set lifts the restriction.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtMost(maxAllowedNamespacePrefixes),
				},
			},
			"managed_by": schema.StringAttribute{
				Description: "The system that manages the space's Fleet settings, as reported by Fleet.",
				Computed:    true,
			},
		},
	}
}
