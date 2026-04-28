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

package importsavedobjects

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *Resource) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		// create_new_copies = true cannot be combined with overwrite = true
		bothTrueConflict("create_new_copies", "overwrite"),
		// create_new_copies = true cannot be combined with compatibility_mode = true
		bothTrueConflict("create_new_copies", "compatibility_mode"),
	}
}

// bothTrueConflict returns a ConfigValidator that errors only when both named
// boolean attributes are explicitly set to true. This is narrower than
// resourcevalidator.Conflicting, which fires whenever both attributes are
// non-null regardless of their values.
func bothTrueConflict(attr1, attr2 string) resource.ConfigValidator {
	return &bothTrueValidator{attr1: attr1, attr2: attr2}
}

type bothTrueValidator struct {
	attr1 string
	attr2 string
}

func (v *bothTrueValidator) Description(_ context.Context) string {
	return fmt.Sprintf("%s and %s cannot both be true", v.attr1, v.attr2)
}

func (v *bothTrueValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *bothTrueValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var val1, val2 types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(v.attr1), &val1)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(v.attr2), &val2)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !val1.IsNull() && !val1.IsUnknown() && val1.ValueBool() &&
		!val2.IsNull() && !val2.IsUnknown() && val2.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root(v.attr1),
			"Invalid attribute combination",
			fmt.Sprintf("%s and %s cannot both be set to true", v.attr1, v.attr2),
		)
	}
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Create sets of Kibana saved objects from a file created by the export API. See https://www.elastic.co/guide/en/kibana/current/saved-objects-api-import.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated ID for the import.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"ignore_import_errors": schema.BoolAttribute{
				Description: "If set to true, errors during the import process will not fail the configuration application",
				Optional:    true,
			},
			"create_new_copies": schema.BoolAttribute{
				Description: "Creates copies of saved objects, regenerates each object ID, and resets the origin. " +
					"When used, potential conflict errors are avoided. Cannot be used with overwrite or compatibility_mode.",
				Optional: true,
			},
			"overwrite": schema.BoolAttribute{
				Description: "Overwrites saved objects when they already exist. When used, potential conflict errors are automatically resolved by overwriting the destination object.",
				Optional:    true,
			},
			"compatibility_mode": schema.BoolAttribute{
				Description: "Applies various adjustments to the saved objects that are being imported to maintain " +
					"compatibility between different Kibana versions. Use this option only if you encounter issues with " +
					"imported saved objects. Cannot be used with create_new_copies.",
				Optional: true,
			},
			"file_contents": schema.StringAttribute{
				Description: "The contents of the exported saved objects file.",
				Required:    true,
			},

			"success": schema.BoolAttribute{
				Description: successDescription,
				Computed:    true,
			},
			"success_count": schema.Int64Attribute{
				Description: "Indicates the number of successfully imported records.",
				Computed:    true,
			},
			"errors": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":    types.StringType,
						"type":  types.StringType,
						"title": types.StringType,
						"error": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"type": types.StringType,
							},
						},
						"meta": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"icon":  types.StringType,
								"title": types.StringType,
							},
						},
					},
				},
			},
			"success_results": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":             types.StringType,
						"type":           types.StringType,
						"destination_id": types.StringType,
						"meta": types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"icon":  types.StringType,
								"title": types.StringType,
							},
						},
					},
				},
			},
		},

		Blocks: map[string]schema.Block{
			"kibana_connection": providerschema.GetKbFWConnectionBlock(),
		}}
}

type Resource struct {
	*entitycore.ResourceBase
}

func newResource() *Resource {
	return &Resource{
		ResourceBase: entitycore.NewResourceBase(entitycore.ComponentKibana, "import_saved_objects"),
	}
}

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                     = newResource()
	_ resource.ResourceWithConfigure        = newResource()
	_ resource.ResourceWithConfigValidators = newResource()
)

// NewResource returns a new Resource instance for provider registration and tests.
func NewResource() resource.Resource {
	return newResource()
}

type modelV0 struct {
	ID                 types.String `tfsdk:"id"`
	KibanaConnection   types.List   `tfsdk:"kibana_connection"`
	SpaceID            types.String `tfsdk:"space_id"`
	IgnoreImportErrors types.Bool   `tfsdk:"ignore_import_errors"`
	CreateNewCopies    types.Bool   `tfsdk:"create_new_copies"`
	Overwrite          types.Bool   `tfsdk:"overwrite"`
	CompatibilityMode  types.Bool   `tfsdk:"compatibility_mode"`
	FileContents       types.String `tfsdk:"file_contents"`
	Success            types.Bool   `tfsdk:"success"`
	SuccessCount       types.Int64  `tfsdk:"success_count"`
	Errors             types.List   `tfsdk:"errors"`
	SuccessResults     types.List   `tfsdk:"success_results"`
}
