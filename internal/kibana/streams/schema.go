package streams

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a Kibana Stream (wired/classic ingest or group) via the Streams API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Internal identifier of the stream (`<cluster_uuid>/<name>`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Stream name (path segment `{name}` in the Streams API).",
				PlanModifiers: []planmodifier.String{
					// Renaming a stream requires creating a new stream in Kibana and
					// deleting the old one; model this as a replacement in Terraform.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
				MarkdownDescription: "Kibana space ID; defaults to `default`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional human‑readable description for the stream.",
			},
			"type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed stream type, e.g. `wired`, `classic`, or `group`.",
			},
			"create_if_missing": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				MarkdownDescription: "When `true` and a `group` block is configured, Terraform will create a new group stream via " +
					"`PUT /api/streams/{name}` if the stream does not already exist. When `false` (default), the stream " +
					"must already exist and Terraform will only manage its group configuration via `/_group`.",
			},
			"ingest": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Ingest settings for wired/classic ingest streams. In the current POC this is populated from `_ingest` and treated as read-only. Only the `type` field is exposed for now.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Ingest type reported by Kibana, e.g. `wired` or `classic`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"group": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Grouping configuration for group streams.",
				Attributes: map[string]schema.Attribute{
					"members": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Member stream names that belong to this group.",
					},
					"metadata": schema.MapAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Arbitrary metadata for the group stream.",
					},
					"tags": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Tags associated with this group stream.",
					},
				},
			},
		},
	}
}
