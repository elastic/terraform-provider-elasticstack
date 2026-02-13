package workflow

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	SpaceID    types.String `tfsdk:"space_id"`
	WorkflowID types.String `tfsdk:"workflow_id"`
	Yaml       types.String `tfsdk:"yaml"`
}
