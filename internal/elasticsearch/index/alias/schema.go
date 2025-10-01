package alias

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *aliasResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Elasticsearch alias. " +
			"See, https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-aliases.html",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Generated ID of the alias resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The alias name.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"write_index": schema.SingleNestedBlock{
				Description: "The write index for the alias. Only one write index is allowed per alias.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description: "Name of the write index.",
						Required:    true,
					},
					"filter": schema.StringAttribute{
						Description: "Query used to limit documents the alias can access.",
						Optional:    true,
						CustomType:  jsontypes.NormalizedType{},
					},
					"index_routing": schema.StringAttribute{
						Description: "Value used to route indexing operations to a specific shard.",
						Optional:    true,
					},
					"is_hidden": schema.BoolAttribute{
						Description: "If true, the alias is hidden.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"routing": schema.StringAttribute{
						Description: "Value used to route indexing and search operations to a specific shard.",
						Optional:    true,
					},
					"search_routing": schema.StringAttribute{
						Description: "Value used to route search operations to a specific shard.",
						Optional:    true,
					},
				},
			},
			"read_indices": schema.SetNestedBlock{
				Description: "Set of read indices for the alias.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the read index.",
							Required:    true,
						},
						"filter": schema.StringAttribute{
							Description: "Query used to limit documents the alias can access.",
							Optional:    true,
							CustomType:  jsontypes.NormalizedType{},
						},
						"index_routing": schema.StringAttribute{
							Description: "Value used to route indexing operations to a specific shard.",
							Optional:    true,
						},
						"is_hidden": schema.BoolAttribute{
							Description: "If true, the alias is hidden.",
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
						},
						"routing": schema.StringAttribute{
							Description: "Value used to route indexing and search operations to a specific shard.",
							Optional:    true,
						},
						"search_routing": schema.StringAttribute{
							Description: "Value used to route search operations to a specific shard.",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}
