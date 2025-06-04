package parameter

import (
	"slices"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
		MarkdownDescription: "Synthetics parameter config, see https://www.elastic.co/docs/api/doc/kibana/group/endpoint-synthetics for more details",
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
				MarkdownDescription: "A description of the parameter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				Optional:            true,
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

func (m *tfModelV0) toParameterConfig(forUpdate bool) kbapi.ParameterConfig {
	// share_across_spaces is not allowed to be set when updating an existing
	// global parameter.
	var shareAcrossSpaces *bool = nil
	if !forUpdate {
		shareAcrossSpaces = m.ShareAcrossSpaces.ValueBoolPointer()
	}

	return kbapi.ParameterConfig{
		Key:               m.Key.ValueString(),
		Value:             m.Value.ValueString(),
		Description:       m.Description.ValueString(),
		Tags:              synthetics.ValueStringSlice(m.Tags),
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

func toModelV0(param kbapi.Parameter) tfModelV0 {
	allSpaces := slices.Equal(param.Namespaces, []string{"*"})

	return tfModelV0{
		ID:                types.StringValue(param.Id),
		Key:               types.StringValue(param.Key),
		Value:             types.StringValue(param.Value),
		Description:       types.StringValue(param.Description),
		Tags:              synthetics.StringSliceValue(param.Tags),
		ShareAcrossSpaces: types.BoolValue(allSpaces),
	}
}
