package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *outputResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = getSchema()
}

func getSchema() schema.Schema {
	return schema.Schema{
		Version:     1,
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
					stringvalidator.OneOf("elasticsearch", "logstash", "kafka"),
				},
			},
			"hosts": schema.ListAttribute{
				Description: "A list of hosts.",
				Required:    true,
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
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this output is available. When set, the output will be created and managed within the specified space. Note: The order of space IDs does not matter as this is a set.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_id": schema.StringAttribute{
				Description: "The ID of the proxy to use for this output. The proxy must exist in the same space as the output.",
				Optional:    true,
			},
			"ssl": schema.SingleNestedAttribute{
				Description: "SSL configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"certificate_authorities": schema.ListAttribute{
						Description: "Server SSL certificate authorities.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					"certificate": schema.StringAttribute{
						Description: "Client SSL certificate.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					"key": schema.StringAttribute{
						Description: "Client SSL certificate key.",
						Required:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
			},
			"kafka": schema.SingleNestedAttribute{
				Description: "Kafka-specific configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"auth_type": schema.StringAttribute{
						Description: "Authentication type for Kafka output.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "user_pass", "ssl", "kerberos"),
						},
					},
					"broker_timeout": schema.Float32Attribute{
						Description: "Kafka broker timeout.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Float32{
							float32planmodifier.UseStateForUnknown(),
						},
					},
					"client_id": schema.StringAttribute{
						Description: "Kafka client ID.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"compression": schema.StringAttribute{
						Description: "Compression type for Kafka output.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("gzip", "snappy", "lz4", "none"),
						},
					},
					"compression_level": schema.Int64Attribute{
						Description: "Compression level for Kafka output.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							validators.AllowedIfDependentPathEquals(path.Root("kafka").AtName("compression"), "gzip"),
						},
					},
					"connection_type": schema.StringAttribute{
						Description: "Connection type for Kafka output.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("plaintext", "encryption"),
							validators.AllowedIfDependentPathEquals(
								path.Root("kafka").AtName("auth_type"),
								"none",
							),
						},
					},
					"topic": schema.StringAttribute{
						Description: "Kafka topic.",
						Optional:    true,
					},
					"partition": schema.StringAttribute{
						Description: "Partition strategy for Kafka output.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("random", "round_robin", "hash"),
						},
					},
					"required_acks": schema.Int64Attribute{
						Description: "Number of acknowledgments required for Kafka output.",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.OneOf(-1, 0, 1),
						},
					},
					"timeout": schema.Float32Attribute{
						Description: "Timeout for Kafka output.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Float32{
							float32planmodifier.UseStateForUnknown(),
						},
					},
					"version": schema.StringAttribute{
						Description: "Kafka version.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"username": schema.StringAttribute{
						Description: "Username for Kafka authentication.",
						Optional:    true,
					},
					"password": schema.StringAttribute{
						Description: "Password for Kafka authentication.",
						Optional:    true,
						Sensitive:   true,
					},
					"key": schema.StringAttribute{
						Description: "Key field for Kafka messages.",
						Optional:    true,
					},
					"headers": schema.ListNestedAttribute{
						Description: "Headers for Kafka messages.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Description: "Header key.",
									Required:    true,
								},
								"value": schema.StringAttribute{
									Description: "Header value.",
									Required:    true,
								},
							},
						},
					},
					"hash": schema.SingleNestedAttribute{
						Description: "Hash configuration for Kafka partition.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"hash": schema.StringAttribute{
								Description: "Hash field.",
								Optional:    true,
							},
							"random": schema.BoolAttribute{
								Description: "Use random hash.",
								Optional:    true,
							},
						},
					},
					"random": schema.SingleNestedAttribute{
						Description: "Random configuration for Kafka partition.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								Description: "Number of events to group.",
								Optional:    true,
							},
						},
					},
					"round_robin": schema.SingleNestedAttribute{
						Description: "Round robin configuration for Kafka partition.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"group_events": schema.Float64Attribute{
								Description: "Number of events to group.",
								Optional:    true,
							},
						},
					},
					"sasl": schema.SingleNestedAttribute{
						Description: "SASL configuration for Kafka authentication.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"mechanism": schema.StringAttribute{
								Description: "SASL mechanism.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getSslAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["ssl"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getHeadersAttrTypes() attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["headers"].GetType().(attr.TypeWithElementType).ElementType()
}

func getHashAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["hash"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getRandomAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["random"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getRoundRobinAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["round_robin"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getSaslAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["sasl"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getKafkaAttrTypes() map[string]attr.Type {
	return getSchema().Attributes["kafka"].(schema.SingleNestedAttribute).GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
