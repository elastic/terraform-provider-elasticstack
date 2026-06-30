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

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Creates a new Fleet Server Host.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
			},
			"host_id": schema.StringAttribute{
				Description: `Unique identifier of the Fleet server host. When omitted, Fleet auto-generates an ID. ` +
					`When set, the value must be 1-255 characters and must not contain path separators ("/"), ` +
					`traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype"). ` +
					`Invalid explicit values fail at plan time.`,
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fleet.IDValidator("host_id"),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Fleet server host.",
				Required:    true,
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
			"default": schema.BoolAttribute{
				Description: "Set as default.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"space_ids": schema.SetAttribute{
				Description: spaceIDsDescription,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}
