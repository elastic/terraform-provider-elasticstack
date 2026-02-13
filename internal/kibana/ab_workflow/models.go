package ab_workflow

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workflowModel struct {
	ID            types.String `tfsdk:"id"`
	Configuration types.String `tfsdk:"configuration"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Valid         types.Bool   `tfsdk:"valid"`
}

func (model *workflowModel) populateFromAPI(data *kbapi.WorkflowDetailDto) diag.Diagnostics {
	if data == nil {
		return nil
	}

	model.ID = types.StringValue(data.Id)
	model.Configuration = types.StringValue(data.Yaml)
	model.Name = types.StringValue(data.Name)

	if data.Description != nil && *data.Description != "" {
		model.Description = types.StringValue(*data.Description)
	} else {
		model.Description = types.StringNull()
	}

	model.Enabled = types.BoolValue(data.Enabled)
	model.Valid = types.BoolValue(data.Valid)

	return nil
}

func (model workflowModel) toAPICreateModel() (kbapi.CreateWorkflowCommand, diag.Diagnostics) {
	body := kbapi.CreateWorkflowCommand{
		Yaml: model.Configuration.ValueString(),
	}

	if !model.ID.IsNull() && !model.ID.IsUnknown() {
		id := model.ID.ValueString()
		body.Id = &id
	}

	return body, nil
}

func (model workflowModel) toAPIUpdateModel() (kbapi.UpdateWorkflowCommand, diag.Diagnostics) {
	yaml := model.Configuration.ValueString()
	body := kbapi.UpdateWorkflowCommand{
		Yaml: &yaml,
	}

	return body, nil
}
