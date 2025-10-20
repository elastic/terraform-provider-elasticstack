package parameter

import (
	_ "embed"
	"slices"
	"strings"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

//go:embed resource-description.md
var syntheticsParameterDescription string

type tfModelV0 struct {
	ID                types.String   `tfsdk:"id"`
	Key               types.String   `tfsdk:"key"`
	Value             types.String   `tfsdk:"value"`
	Description       types.String   `tfsdk:"description"`
	Tags              []types.String `tfsdk:"tags"` //> string
	ShareAcrossSpaces types.Bool     `tfsdk:"share_across_spaces"`
}

func parameterSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: syntheticsParameterDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Generated id for the parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				MarkdownDescription: "The key of the parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"value": schema.StringAttribute{
				Optional:            false,
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The value associated with the parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "A description of the parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				MarkdownDescription: "An array of tags to categorize the parameter.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"share_across_spaces": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the parameter should be shared across spaces.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (m *tfModelV0) toParameterRequest(forUpdate bool) kboapi.SyntheticsParameterRequest {
	// share_across_spaces is not allowed to be set when updating an existing
	// global parameter.
	var shareAcrossSpaces *bool = nil
	if !forUpdate {
		shareAcrossSpaces = m.ShareAcrossSpaces.ValueBoolPointer()
	}

	return kboapi.SyntheticsParameterRequest{
		Key:         m.Key.ValueString(),
		Value:       m.Value.ValueString(),
		Description: utils.Pointer(m.Description.ValueString()),
		// We need this to marshal as an empty JSON array, not null.
		Tags:              utils.Pointer(utils.NonNilSlice(synthetics.ValueStringSlice(m.Tags))),
		ShareAcrossSpaces: shareAcrossSpaces,
	}
}

func tryReadCompositeId(id string) (*clients.CompositeId, diag.Diagnostics) {
	if strings.Contains(id, "/") {
		compositeId, diagnostics := synthetics.GetCompositeId(id)
		return compositeId, diagnostics
	}
	return nil, diag.Diagnostics{}
}

func modelV0FromOAPI(param kboapi.SyntheticsGetParameterResponse) tfModelV0 {
	allSpaces := slices.Equal(*param.Namespaces, []string{"*"})

	return tfModelV0{
		ID:          types.StringPointerValue(param.Id),
		Key:         types.StringPointerValue(param.Key),
		Value:       types.StringPointerValue(param.Value),
		Description: types.StringPointerValue(param.Description),
		// Terraform, like json.Marshal, treats empty slices as null. We need an
		// actual backing array of size 0.
		Tags:              utils.NonNilSlice(synthetics.StringSliceValue(utils.DefaultIfNil(param.Tags))),
		ShareAcrossSpaces: types.BoolValue(allSpaces),
	}
}
