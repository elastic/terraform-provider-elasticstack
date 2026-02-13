package tool

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ToolID        types.String `tfsdk:"tool_id"`
	Type          types.String `tfsdk:"type"`
	Description   types.String `tfsdk:"description"`
	Tags          types.List   `tfsdk:"tags"`
	ReadOnly      types.Bool   `tfsdk:"readonly"`
	Configuration types.String `tfsdk:"configuration"`
}
