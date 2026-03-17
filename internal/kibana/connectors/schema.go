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

package connectors

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Creates a Kibana action connector. See https://www.elastic.co/guide/en/kibana/current/action-types.html",
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connector_id": schema.StringAttribute{
				Description: "A UUID v1 or v4 to use instead of a randomly generated ID.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					validators.IsUUID(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the connector. While this name does not have to be unique, a distinctive name can help you identify a connector.",
				Required:    true,
			},
			"connector_type_id": schema.StringAttribute{
				Description: "The ID of the connector type, e.g. `.index`.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				CustomType:  NewConfigType(),
				Description: customtypes.DescriptionWithContextWarning("The configuration for the connector. Configuration properties vary depending on the connector type."),
				Optional:    true,
				Computed:    true,
			},
			"secrets": schema.StringAttribute{
				CustomType:  jsontypes.NormalizedType{},
				Description: "The secrets configuration for the connector. Secrets configuration properties vary depending on the connector type.",
				Optional:    true,
				Sensitive:   true,
			},
			"is_deprecated": schema.BoolAttribute{
				Description: "Indicates whether the connector type is deprecated.",
				Computed:    true,
			},
			"is_missing_secrets": schema.BoolAttribute{
				Description: "Indicates whether secrets are missing for the connector.",
				Computed:    true,
			},
			"is_preconfigured": schema.BoolAttribute{
				Description: "Indicates whether it is a preconfigured connector.",
				Computed:    true,
			},
		},
	}
}
