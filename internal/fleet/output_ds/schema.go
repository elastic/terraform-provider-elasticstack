package output_ds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *outputDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns information about a Fleet output. See the [Fleet output API documentation](https://www.elastic.co/docs/api/doc/kibana/v9/group/endpoint-fleet-outputs) for more details.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "The name of the output.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "The Kibana space ID where this output is available.",
				Optional:    true,
			},
			"output_id": schema.StringAttribute{
				Description: "Unique identifier of the output.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The output type.",
				Computed:    true,
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"ca_sha256": schema.StringAttribute{
				Description: "Fingerprint of the Elasticsearch CA certificate.",
				Computed:    true,
			},
			"ca_trusted_fingerprint": schema.StringAttribute{
				Description: "Fingerprint of trusted CA.",
				Computed:    true,
			},
			"default_integrations": schema.BoolAttribute{
				Description: "This output is the default for agent integrations.",
				Computed:    true,
			},
			"default_monitoring": schema.BoolAttribute{
				Description: "This output is the default for agent monitoring.",
				Computed:    true,
			},
			"config_yaml": schema.StringAttribute{
				Description: "Advanced YAML configuration.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}
