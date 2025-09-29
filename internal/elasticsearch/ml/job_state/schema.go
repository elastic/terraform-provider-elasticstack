package job_state

import (
	"context"
	_ "embed"
	"regexp"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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

//go:embed resource-description.md
var mlJobStateResourceDescription string

func (r *mlJobStateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: mlJobStateResourceDescription,
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
			"job_id": schema.StringAttribute{
				MarkdownDescription: "Identifier for the anomaly detection job.",
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
				MarkdownDescription: "The desired state for the ML job. Valid values are `opened` and `closed`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("opened", "closed"),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "When closing a job, use to forcefully close it. This method is quicker but can miss important clean up tasks.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"job_timeout": schema.StringAttribute{
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
