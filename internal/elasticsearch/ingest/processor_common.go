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

package ingest

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CommonProcessorModel holds the fields shared by every ingest processor data
// source. Embed this struct in concrete processor models.
type CommonProcessorModel struct {
	Description   types.String `tfsdk:"description"`
	If            types.String `tfsdk:"if"`
	IgnoreFailure types.Bool   `tfsdk:"ignore_failure"`
	OnFailure     types.List   `tfsdk:"on_failure"`
	Tag           types.String `tfsdk:"tag"`
}

// CommonProcessorSchemaAttributes returns the schema attributes that are common
// to all ingest processor data sources.
func CommonProcessorSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			Description: "Description of the processor.",
			Optional:    true,
		},
		"if": schema.StringAttribute{
			Description: "Conditionally execute the processor",
			Optional:    true,
		},
		"ignore_failure": schema.BoolAttribute{
			Description: "Ignore failures for the processor.",
			Optional:    true,
			Computed:    true,
		},
		"on_failure": schema.ListAttribute{
			Description: "Handle failures for the processor.",
			Optional:    true,
			ElementType: jsontypes.NormalizedType{},
			Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
		},
		"tag": schema.StringAttribute{
			Description: "Identifier for the processor.",
			Optional:    true,
		},
	}
}
