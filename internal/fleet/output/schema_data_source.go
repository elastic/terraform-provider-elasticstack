package output

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Data source for Fleet outputs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier in the format `<space_id>/<output_id>`.",
				Computed:            true,
			},
			"output_id": schema.StringAttribute{
				MarkdownDescription: "The Fleet output ID.",
				Required:            true,
			},
			"space_id": schema.StringAttribute{
				MarkdownDescription: "An identifier for the space. If space_id is not provided, the default space is used.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the output.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The output type.",
				Computed:            true,
			},
			"hosts": schema.ListAttribute{
				MarkdownDescription: "A list of hosts.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"ca_sha256": schema.StringAttribute{
				MarkdownDescription: "Fingerprint of the Elasticsearch CA certificate.",
				Computed:            true,
			},
			"ca_trusted_fingerprint": schema.StringAttribute{
				MarkdownDescription: "Fingerprint of trusted CA.",
				Computed:            true,
			},
			"default_integrations": schema.BoolAttribute{
				MarkdownDescription: "Make this output the default for agent integrations.",
				Computed:            true,
			},
			"default_monitoring": schema.BoolAttribute{
				MarkdownDescription: "Make this output the default for agent monitoring.",
				Computed:            true,
			},
			"config_yaml": schema.StringAttribute{
				MarkdownDescription: "Advanced YAML configuration returned by the API.",
				Computed:            true,
				Sensitive:           true,
			},
			"ssl": schema.SingleNestedAttribute{
				MarkdownDescription: "SSL configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"certificate_authorities": schema.ListAttribute{
						MarkdownDescription: "Server SSL certificate authorities.",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"certificate": schema.StringAttribute{
						MarkdownDescription: "Client SSL certificate.",
						Computed:            true,
					},
					"key": schema.StringAttribute{
						MarkdownDescription: "Client SSL certificate key.",
						Computed:            true,
						Sensitive:           true,
					},
				},
			},
			"kafka": schema.SingleNestedAttribute{
				MarkdownDescription: "Kafka-specific configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						MarkdownDescription: "Authentication type for Kafka output.",
						Computed:            true,
					},
					"broker_timeout": schema.Float32Attribute{
						MarkdownDescription: "Kafka broker timeout.",
						Computed:            true,
					},
					"client_id": schema.StringAttribute{
						MarkdownDescription: "Kafka client ID.",
						Computed:            true,
					},
					"compression": schema.StringAttribute{
						MarkdownDescription: "Compression type for Kafka output.",
						Computed:            true,
					},
					"compression_level": schema.Int64Attribute{
						MarkdownDescription: "Compression level for Kafka output.",
						Computed:            true,
					},
					"connection_type": schema.StringAttribute{
						MarkdownDescription: "Connection type for Kafka output.",
						Computed:            true,
					},
					"topic": schema.StringAttribute{
						MarkdownDescription: "Kafka topic.",
						Computed:            true,
					},
					"partition": schema.StringAttribute{
						MarkdownDescription: "Partition strategy for Kafka output.",
						Computed:            true,
					},
					"required_acks": schema.Int64Attribute{
						MarkdownDescription: "Number of acknowledgments required for Kafka output.",
						Computed:            true,
					},
					"timeout": schema.Float32Attribute{
						MarkdownDescription: "Timeout for Kafka output.",
						Computed:            true,
					},
					"version": schema.StringAttribute{
						MarkdownDescription: "Kafka version.",
						Computed:            true,
					},
					"username": schema.StringAttribute{
						MarkdownDescription: "Username for Kafka authentication.",
						Computed:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "Password for Kafka authentication.",
						Computed:            true,
						Sensitive:           true,
					},
					"key": schema.StringAttribute{
						MarkdownDescription: "Key field for Kafka messages.",
						Computed:            true,
					},
					"headers": schema.ListNestedAttribute{
						MarkdownDescription: "Headers for Kafka messages.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									MarkdownDescription: "Header key.",
									Computed:            true,
								},
								"value": schema.StringAttribute{
									MarkdownDescription: "Header value.",
									Computed:            true,
								},
							},
						},
					},
					"hash": schema.SingleNestedAttribute{
						MarkdownDescription: "Hash configuration for Kafka partition.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"hash": schema.StringAttribute{
								MarkdownDescription: "Hash field.",
								Computed:            true,
							},
							"random": schema.BoolAttribute{
								MarkdownDescription: "Use random hash.",
								Computed:            true,
							},
						},
					},
					"random": schema.SingleNestedAttribute{
						MarkdownDescription: "Random configuration for Kafka partition.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								MarkdownDescription: "Number of events to group.",
								Computed:            true,
							},
						},
					},
					"round_robin": schema.SingleNestedAttribute{
						MarkdownDescription: "Round robin configuration for Kafka partition.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								MarkdownDescription: "Number of events to group.",
								Computed:            true,
							},
						},
					},
					"sasl": schema.SingleNestedAttribute{
						MarkdownDescription: "SASL configuration for Kafka authentication.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"mechanism": schema.StringAttribute{
								MarkdownDescription: "SASL mechanism.",
								Computed:            true,
							},
						},
					},
				},
			},
		},
	}
}
