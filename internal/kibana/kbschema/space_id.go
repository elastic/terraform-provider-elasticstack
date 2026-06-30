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

package kbschema

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
)

const spaceIDDescription = "An identifier for the space. If space_id is not provided, the default space is used."

// ResourceSpaceIDAttribute returns the canonical space_id attribute for Kibana
// resources that support UseStateForUnknown in addition to RequiresReplace.
func ResourceSpaceIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: spaceIDDescription,
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(clients.DefaultSpaceID),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// ResourceSpaceIDAttributeRequiresReplaceOnly returns the canonical space_id
// attribute for Kibana resources that only need RequiresReplace (no
// UseStateForUnknown).
func ResourceSpaceIDAttributeRequiresReplaceOnly() schema.StringAttribute {
	return schema.StringAttribute{
		MarkdownDescription: spaceIDDescription,
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(clients.DefaultSpaceID),
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// DataSourceSpaceIDAttribute returns the canonical space_id attribute for
// Kibana data sources (Optional+Computed, no plan modifiers).
func DataSourceSpaceIDAttribute() dsschema.StringAttribute {
	return dsschema.StringAttribute{
		MarkdownDescription: spaceIDDescription,
		Optional:            true,
		Computed:            true,
	}
}
