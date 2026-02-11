package output

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (d *outputDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = getDataSourceSchema()
}

func getDataSourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Retrieves a Fleet Output.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this data source.",
				Computed:    true,
			},
			"output_id": schema.StringAttribute{
				Description: "Unique identifier of the output.",
				Required:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "The Kibana space ID to query the output from. If not specified, queries the default space.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the output.",
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
				Description: "Indicates if this output is the default for agent integrations.",
				Computed:    true,
			},
			"default_monitoring": schema.BoolAttribute{
				Description: "Indicates if this output is the default for agent monitoring.",
				Computed:    true,
			},
			"config_yaml": schema.StringAttribute{
				Description: "Advanced YAML configuration. YAML settings here will be added to the output section of each agent policy.",
				Computed:    true,
				Sensitive:   true,
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this output is available. Note: This value is not returned by the Fleet API.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"ssl": schema.SingleNestedAttribute{
				Description: "SSL configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"certificate_authorities": schema.ListAttribute{
						Description: "Server SSL certificate authorities.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"certificate": schema.StringAttribute{
						Description: "Client SSL certificate.",
						Computed:    true,
					},
					"key": schema.StringAttribute{
						Description: "Client SSL certificate key.",
						Computed:    true,
						Sensitive:   true,
					},
				},
			},
			"kafka": schema.SingleNestedAttribute{
				Description: "Kafka-specific configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						Description: "Authentication type for Kafka output.",
						Computed:    true,
					},
					"broker_timeout": schema.Float32Attribute{
						Description: "Kafka broker timeout.",
						Computed:    true,
					},
					"client_id": schema.StringAttribute{
						Description: "Kafka client ID.",
						Computed:    true,
					},
					"compression": schema.StringAttribute{
						Description: "Compression type for Kafka output.",
						Computed:    true,
					},
					"compression_level": schema.Int64Attribute{
						Description: "Compression level for Kafka output.",
						Computed:    true,
					},
					"connection_type": schema.StringAttribute{
						Description: "Connection type for Kafka output.",
						Computed:    true,
					},
					"topic": schema.StringAttribute{
						Description: "Kafka topic.",
						Computed:    true,
					},
					"partition": schema.StringAttribute{
						Description: "Partition strategy for Kafka output.",
						Computed:    true,
					},
					"required_acks": schema.Int64Attribute{
						Description: "Number of acknowledgments required for Kafka output.",
						Computed:    true,
					},
					"timeout": schema.Float32Attribute{
						Description: "Timeout for Kafka output.",
						Computed:    true,
					},
					"version": schema.StringAttribute{
						Description: "Kafka version.",
						Computed:    true,
					},
					"username": schema.StringAttribute{
						Description: "Username for Kafka authentication.",
						Computed:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password for Kafka authentication.",
						Computed:    true,
						Sensitive:   true,
					},
					"key": schema.StringAttribute{
						Description: "Key field for Kafka messages.",
						Computed:    true,
					},
					"headers": schema.ListNestedAttribute{
						Description: "Headers for Kafka messages.",
						Computed:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Description: "Header key.",
									Computed:    true,
								},
								"value": schema.StringAttribute{
									Description: "Header value.",
									Computed:    true,
								},
							},
						},
					},
					"hash": schema.SingleNestedAttribute{
						Description: "Hash configuration for Kafka partition.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"hash": schema.StringAttribute{
								Description: "Hash field.",
								Computed:    true,
							},
							"random": schema.BoolAttribute{
								Description: "Use random hash.",
								Computed:    true,
							},
						},
					},
					"random": schema.SingleNestedAttribute{
						Description: "Random configuration for Kafka partition.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								Description: "Number of events to group.",
								Computed:    true,
							},
						},
					},
					"round_robin": schema.SingleNestedAttribute{
						Description: "Round robin configuration for Kafka partition.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								Description: "Number of events to group.",
								Computed:    true,
							},
						},
					},
					"sasl": schema.SingleNestedAttribute{
						Description: "SASL configuration for Kafka authentication.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"mechanism": schema.StringAttribute{
								Description: "SASL mechanism.",
								Computed:    true,
							},
						},
					},
				},
			},
		},
	}
}
