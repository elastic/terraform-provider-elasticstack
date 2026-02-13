package agent

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema defines the schema for the data source.
func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Export an Agent Builder agent by ID, optionally including its tools and workflows. See https://www.elastic.co/guide/en/kibana/current/agent-builder-api.html",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The agent ID to export.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:    true,
			},
			"include_dependencies": schema.BoolAttribute{
				Description: "If true, exports the agent along with its tools and workflows. If false (default), exports only the agent.",
				Optional:    true,
			},
			"agent": schema.StringAttribute{
				Description: "The exported agent in JSON format.",
				Computed:    true,
			},
			"tools": schema.ListNestedAttribute{
				Description: "List of exported tools (only populated when include_dependencies is true).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The tool ID.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the tool (esql, index_search, workflow, mcp).",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of what the tool does.",
							Computed:    true,
						},
						"tags": schema.ListAttribute{
							Description: "Tags for categorizing and organizing tools.",
							ElementType: types.StringType,
							Computed:    true,
						},
						"readonly": schema.BoolAttribute{
							Description: "Whether the tool is read-only.",
							Computed:    true,
						},
						"configuration": schema.StringAttribute{
							Description: "The tool configuration in JSON format.",
							Computed:    true,
						},
					},
				},
			},
			"workflows": schema.ListNestedAttribute{
				Description: "List of exported workflows (only populated when include_dependencies is true).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The workflow ID.",
							Computed:    true,
						},
						"yaml": schema.StringAttribute{
							Description: "The workflow definition in YAML format.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}
