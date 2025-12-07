package default_data_view

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type defaultDataViewModel struct {
	ID         types.String `tfsdk:"id"`
	DataViewID types.String `tfsdk:"data_view_id"`
	Force      types.Bool   `tfsdk:"force"`
	SkipDelete types.Bool   `tfsdk:"skip_delete"`
	SpaceID    types.String `tfsdk:"space_id"`
}
