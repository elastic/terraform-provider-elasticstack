package workflow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Schema defines the schema for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export an Agent Builder workflow by ID. See https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The workflow ID to export.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"workflow_id": schema.StringAttribute{
				Description: "The ID of the exported workflow.",
				Computed:    true,
			},
			"yaml": schema.StringAttribute{
				Description: "The exported workflow definition in YAML format.",
				Computed:    true,
			},
		},
	}
}
