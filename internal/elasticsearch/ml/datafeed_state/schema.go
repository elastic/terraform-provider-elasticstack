package datafeed_state

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/ml/datafeed"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *mlDatafeedStateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

//go:embed resource-description.md
var description string

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: description,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock("elasticsearch_connection", false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"datafeed_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the ML datafeed.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain only alphanumeric characters, hyphens, and underscores"),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The desired state for the ML datafeed. Valid values are `started` and `stopped`.",
				Required:            true,
				Validators: []validator.String{
					// We don't allow starting/stopping here since they're transient states
					stringvalidator.OneOf(string(datafeed.StateStarted), string(datafeed.StateStopped)),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "When stopping a datafeed, use to forcefully stop it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"start": schema.StringAttribute{
				MarkdownDescription: "The time that the datafeed should start collecting data. When not specified, the datafeed starts in real-time. This property must be specified in RFC 3339 format.",
				CustomType:          timetypes.RFC3339Type{},
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end": schema.StringAttribute{
				MarkdownDescription: "The time that the datafeed should end collecting data. When not specified, the datafeed continues in real-time. This property must be specified in RFC 3339 format.",
				CustomType:          timetypes.RFC3339Type{},
				Optional:            true,
			},
			"datafeed_timeout": schema.StringAttribute{
				MarkdownDescription: "Timeout for the operation. Examples: `30s`, `5m`, `1h`. Default is `30s`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("30s"),
				CustomType:          customtypes.DurationType{},
			},
			"timeouts": timeouts.Attributes(context.Background(), timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}
