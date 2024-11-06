package data_stream_lifecycle

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Configures the data stream lifecycle for the targeted data streams, see: https://www.elastic.co/guide/en/elasticsearch/reference/current/data-stream-apis.html",
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier of the resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the data stream. Supports wildcards.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_retention": schema.StringAttribute{
				Description: "Every document added to this data stream will be stored at least for this time frame. When empty, every document in this data stream will be stored indefinitely",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Data stream lifecycle on/off.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"expand_wildcards": schema.StringAttribute{
				Description: "Determines how wildcard patterns in the `indices` parameter match data streams and indices. Supports comma-separated values, such as `closed,hidden`.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("open"),
				Validators: []validator.String{
					stringvalidator.OneOf("all", "open", "closed", "hidden", "none"),
				},
			},
			"downsampling": schema.ListNestedAttribute{
				Description: "Downsampling configuration objects, each defining an after interval representing when the backing index is meant to be downsampled and a fixed_interval representing the downsampling interval.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"after": schema.StringAttribute{
							Description: "Interval representing when the backing index is meant to be downsampled",
							Required:    true,
						},
						"fixed_interval": schema.StringAttribute{
							Description: "The interval at which to aggregate the original time series index.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func downsamplingElementType() attr.Type {
	return getSchema().Attributes["downsampling"].GetType().(attr.TypeWithElementType).ElementType()
}
