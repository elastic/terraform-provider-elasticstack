package integration_ds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func (d *integrationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema.Description = `This data source provides information about a Fleet integration package. Currently,
the data source will retrieve the latest available version of the package. Version
selection is determined by the Fleet API, which is currently based on semantic
versioning.

By default, the highest GA release version will be selected. If a
package is not GA (the version is below 1.0.0) or if a new non-GA version of the
package is to be selected (i.e., the GA version of the package is 1.5.0, but there's
a new 1.5.1-beta version available), then the ` + "`prerelease`" + ` parameter in the plan
should be set to ` + "`true`."
	resp.Schema.Attributes = map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The ID of this resource.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The integration package name.",
			Required:    true,
		},
		"prerelease": schema.BoolAttribute{
			Description: "Include prerelease packages.",
			Optional:    true,
		},
		"version": schema.StringAttribute{
			Description: "The integration package version.",
			Computed:    true,
		},
	}
}
