package ab_tool

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type toolModel struct {
	ID            types.String `tfsdk:"id"`
	Type          types.String `tfsdk:"type"`
	Description   types.String `tfsdk:"description"`
	Tags          types.List   `tfsdk:"tags"`
	Configuration types.String `tfsdk:"configuration"`
}

func (model *toolModel) populateFromAPI(ctx context.Context, data *models.Tool) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue(data.ID)
	model.Type = types.StringValue(data.Type)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(data.Tags) > 0 {
		tags, d := types.ListValueFrom(ctx, types.StringType, data.Tags)
		diags.Append(d...)
		model.Tags = tags
	} else {
		model.Tags = types.ListNull(types.StringType)
	}

	if data.Configuration != nil {
		configJSON, err := json.Marshal(data.Configuration)
		if err != nil {
			diags.AddError("Configuration Error", "Failed to marshal configuration to JSON: "+err.Error())
			return diags
		}
		model.Configuration = types.StringValue(string(configJSON))
	} else {
		model.Configuration = types.StringNull()
	}

	return diags
}

func (model toolModel) toAPICreateModel(ctx context.Context) (kbapi.PostAgentBuilderToolsJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration := make(map[string]interface{})

	if !model.Configuration.IsNull() && model.Configuration.ValueString() != "" {
		if err := json.Unmarshal([]byte(model.Configuration.ValueString()), &configuration); err != nil {
			diags.AddError("Configuration Error", "Failed to parse configuration JSON: "+err.Error())
			return kbapi.PostAgentBuilderToolsJSONRequestBody{}, diags
		}
	}

	body := kbapi.PostAgentBuilderToolsJSONRequestBody{
		Id:            model.ID.ValueString(),
		Type:          kbapi.PostAgentBuilderToolsJSONBodyType(model.Type.ValueString()),
		Configuration: configuration,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	if !model.Tags.IsNull() {
		var tags []string
		d := model.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(d...)
		if len(tags) > 0 {
			body.Tags = &tags
		}
	}

	return body, diags
}

func (model toolModel) toAPIUpdateModel(ctx context.Context) (kbapi.PutAgentBuilderToolsToolidJSONRequestBody, diag.Diagnostics) {
	var diags diag.Diagnostics

	configuration := make(map[string]interface{})

	if !model.Configuration.IsNull() && model.Configuration.ValueString() != "" {
		if err := json.Unmarshal([]byte(model.Configuration.ValueString()), &configuration); err != nil {
			diags.AddError("Configuration Error", "Failed to parse configuration JSON: "+err.Error())
			return kbapi.PutAgentBuilderToolsToolidJSONRequestBody{}, diags
		}
	}

	body := kbapi.PutAgentBuilderToolsToolidJSONRequestBody{
		Configuration: &configuration,
	}

	if !model.Description.IsNull() {
		desc := model.Description.ValueString()
		body.Description = &desc
	}

	if !model.Tags.IsNull() {
		var tags []string
		d := model.Tags.ElementsAs(ctx, &tags, false)
		diags.Append(d...)
		if len(tags) > 0 {
			body.Tags = &tags
		}
	}

	return body, diags
}
