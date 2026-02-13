package agent

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/export_ab"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID                  types.String              `tfsdk:"id"`
	SpaceID             types.String              `tfsdk:"space_id"`
	IncludeDependencies types.Bool                `tfsdk:"include_dependencies"`
	Agent               types.String              `tfsdk:"agent"`
	Tools               []export_ab.ToolModel     `tfsdk:"tools"`
	Workflows           []export_ab.WorkflowModel `tfsdk:"workflows"`
}
