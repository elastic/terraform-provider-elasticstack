package spaces

import "github.com/hashicorp/terraform-plugin-framework/types"

// spacesDataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Spaces []model      `tfsdk:"spaces"`
}

// spacesModel maps coffees schema data.
type model struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	DisabledFeatures types.List   `tfsdk:"disabled_features"`
	Initials         types.String `tfsdk:"initials"`
	Color            types.String `tfsdk:"color"`
	ImageUrl         types.String `tfsdk:"image_url"`
}
