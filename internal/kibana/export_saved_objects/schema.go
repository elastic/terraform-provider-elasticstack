package export_saved_objects

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *dataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export Kibana saved objects. This data source allows you to export saved objects from Kibana and store the result in the Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID for the export.",
				Computed:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"objects": schema.StringAttribute{
				Description: "JSON-encoded list of objects to export. Each object should have 'type' and 'id' fields.",
				Required:    true,
			},
			"exclude_export_details": schema.BoolAttribute{
				Description: "Do not add export details. Defaults to true.",
				Optional:    true,
			},
			"include_references_deep": schema.BoolAttribute{
				Description: "Include references to other saved objects recursively. Defaults to true.",
				Optional:    true,
			},
			"exported_objects": schema.StringAttribute{
				Description: "The exported objects in NDJSON format.",
				Computed:    true,
			},
		},
	}
}

// dataSourceModel maps the data source schema data.
type dataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	SpaceID               types.String `tfsdk:"space_id"`
	Objects               types.String `tfsdk:"objects"`
	ExcludeExportDetails  types.Bool   `tfsdk:"exclude_export_details"`
	IncludeReferencesDeep types.Bool   `tfsdk:"include_references_deep"`
	ExportedObjects       types.String `tfsdk:"exported_objects"`
}
