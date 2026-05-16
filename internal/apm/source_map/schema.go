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

package sourcemap

import (
	"context"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func (r *resourceSourceMap) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Uploads and manages an APM source map artifact. Source maps allow APM to un-minify JavaScript stack traces.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Fleet artifact ID returned by the APM source map upload API. Used to track the resource across plan/apply cycles.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bundle_filepath": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The absolute path of the final bundle as used in the web application (e.g. `/static/js/main.chunk.js`). Must match the path used during the build.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the APM service that the source map applies to. Must match the `service.name` field in APM events.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The version of the APM service that the source map applies to. Must match the `service.version` field in APM events.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sourcemap": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The source map content. Exactly one of `json`, `binary`, or `file.path` must be set.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"json": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "The source map content as a JSON string. Exactly one of `json`, `binary`, or `file.path` must be set. The value is write-only and is not read back from the API.",
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(
								path.MatchRelative().AtParent().AtName("binary"),
								path.MatchRelative().AtParent().AtName("file").AtName("path"),
							),
							stringvalidator.LengthAtLeast(1),
						},
					},
					"binary": schema.StringAttribute{
						Optional:  true,
						Sensitive: true,
						MarkdownDescription: "The source map content as a base64-encoded string (standard encoding). " +
							"Exactly one of `json`, `binary`, or `file.path` must be set. The value is write-only and is not read back from the API.",
						Validators: []validator.String{
							stringvalidator.ExactlyOneOf(
								path.MatchRelative().AtParent().AtName("json"),
								path.MatchRelative().AtParent().AtName("file").AtName("path"),
							),
							stringvalidator.LengthAtLeast(1),
						},
					},
					"file": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "Upload a source map from a local file path.",
						Attributes: map[string]schema.Attribute{
							"path": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Absolute or relative path to the source map file on the local filesystem.",
							},
							"checksum": schema.StringAttribute{
								Computed:            true,
								MarkdownDescription: "SHA256 hex digest of the uploaded sourcemap.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
					},
				},
			},
			"space_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Kibana space ID in which to manage the source map. Omit or set to `\"default\"` for the default space. When set, all API operations are prefixed with `/s/{space_id}`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		},
	}
}
