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

package osquery

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

// ECSMappingSchema returns the resource schema attribute for ecs_mapping.
func ECSMappingSchema() rsschema.MapNestedAttribute {
	return rsschema.MapNestedAttribute{
		MarkdownDescription: "Maps query result columns to ECS field paths. Each map value must set exactly one of `field`, `value`, or `values`.",
		Optional:            true,
		NestedObject: rsschema.NestedAttributeObject{
			Validators: []validator.Object{
				ECSMappingExactlyOneOfValidator(),
			},
			Attributes: map[string]rsschema.Attribute{
				AttrECSMappingField: rsschema.StringAttribute{
					MarkdownDescription: "Query result column name to map from.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				AttrECSMappingValue: rsschema.StringAttribute{
					MarkdownDescription: "Static scalar ECS mapping value.",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.LengthAtLeast(1),
					},
				},
				AttrECSMappingValues: rsschema.SetAttribute{
					MarkdownDescription: "Static array ECS mapping values.",
					Optional:            true,
					ElementType:         types.StringType,
					Validators: []validator.Set{
						setvalidator.SizeAtLeast(1),
					},
				},
			},
		},
	}
}

// ECSMappingDataSourceSchema returns the data source schema attribute for ecs_mapping.
func ECSMappingDataSourceSchema() dsschema.MapNestedAttribute {
	return dsschema.MapNestedAttribute{
		MarkdownDescription: "Maps query result columns to ECS field paths.",
		Computed:            true,
		NestedObject: dsschema.NestedAttributeObject{
			Attributes: map[string]dsschema.Attribute{
				AttrECSMappingField: dsschema.StringAttribute{
					MarkdownDescription: "Query result column name to map from.",
					Computed:            true,
				},
				AttrECSMappingValue: dsschema.StringAttribute{
					MarkdownDescription: "Static scalar ECS mapping value.",
					Computed:            true,
				},
				AttrECSMappingValues: dsschema.SetAttribute{
					MarkdownDescription: "Static array ECS mapping values.",
					Computed:            true,
					ElementType:         types.StringType,
				},
			},
		},
	}
}
