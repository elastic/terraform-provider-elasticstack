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

package systemuser

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

//go:embed resource-description.md
var systemUserResourceDescription string

func (r *systemUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	const usernameAllowedCharsError = "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols " +
		"in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"

	return schema.Schema{
		MarkdownDescription: systemUserResourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "An identifier for the system user (see the [built-in users documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html)).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[[:graph:]]+$`),
						usernameAllowedCharsError,
					),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The user's password. Passwords must be at least 6 characters long.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
					stringvalidator.ConflictsWith(path.MatchRoot("password_hash")),
				},
			},
			"password_hash": schema.StringAttribute{
				MarkdownDescription: systemUserPasswordHashDescription,
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
					stringvalidator.ConflictsWith(path.MatchRoot("password")),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the user is enabled. The default value is true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}
