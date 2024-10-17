package output

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *outputResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Description: "Creates a new Fleet Output.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"output_id": schema.StringAttribute{
				Description: "Unique identifier of the output.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the output.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The output type.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("elasticsearch", "logstash"),
				},
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Optional:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				ElementType: types.StringType,
			},
			"ca_sha256": schema.StringAttribute{
				Description: "Fingerprint of the Elasticsearch CA certificate.",
				Optional:    true,
			},
			"ca_trusted_fingerprint": schema.StringAttribute{
				Description: "Fingerprint of trusted CA.",
				Optional:    true,
			},
			"default_integrations": schema.BoolAttribute{
				Description: "Make this output the default for agent integrations.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"default_monitoring": schema.BoolAttribute{
				Description: "Make this output the default for agent monitoring.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(false),
			},
			"config_yaml": schema.StringAttribute{
				Description: "Advanced YAML configuration. YAML settings here will be added to the output section of each agent policy.",
				Optional:    true,
				Sensitive:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"ssl": schema.ListNestedBlock{
				Description: "SSL configuration.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"certificate_authorities": schema.ListAttribute{
							Description: "Server SSL certificate authorities.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"certificate": schema.StringAttribute{
							Description: "Client SSL certificate.",
							Required:    true,
						},
						"key": schema.StringAttribute{
							Description: "Client SSL certificate key.",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
		},
	}
}

func getSslAttrTypes() attr.Type {
	return getSchema().Blocks["ssl"].Type().(attr.TypeWithElementType).ElementType()
}
