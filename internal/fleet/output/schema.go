// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func getSchema(_ context.Context) schema.Schema {
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
				Description: `Unique identifier of the output. When omitted, Fleet auto-generates an ID. ` +
					`When set, the value must be 1-255 characters and must not contain path separators ("/"), ` +
					`traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype"). ` +
					`Invalid explicit values fail at plan time.`,
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					fleet.IDValidator("output_id"),
				},
			},
			attrName: schema.StringAttribute{
				Description: "The name of the output.",
				Required:    true,
			},
			attrType: schema.StringAttribute{
				Description: "The output type.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(outputTypeElasticsearch, outputTypeLogstash, outputTypeKafka, outputTypeRemoteElasticsearch),
				},
			},
			attrHosts: schema.ListAttribute{
				Description: "A list of hosts.",
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				ElementType: types.StringType,
			},
			"service_token": schema.StringAttribute{
				Description: "Service token for remote Elasticsearch outputs.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					validators.RequiredIfDependentPathEquals(path.Root("type"), "remote_elasticsearch"),
					validators.AllowedIfDependentPathEquals(path.Root("type"), "remote_elasticsearch", validators.AllowedIfOptions{}),
				},
			},
			"sync_integrations": schema.BoolAttribute{
				Description: "When type is remote_elasticsearch, whether Fleet synchronizes integration assets to the remote cluster. Subscription and version requirements apply per Elastic documentation.",
				Optional:    true,
				Validators: []validator.Bool{
					validators.AllowedIfDependentPathEquals(path.Root("type"), "remote_elasticsearch", validators.AllowedIfOptions{}),
				},
			},
			"sync_uninstalled_integrations": schema.BoolAttribute{
				Description: "When type is remote_elasticsearch, whether to sync uninstalled integrations. Only meaningful when sync_integrations is enabled.",
				Optional:    true,
				Validators: []validator.Bool{
					validators.AllowedIfDependentPathEquals(path.Root("type"), "remote_elasticsearch", validators.AllowedIfOptions{}),
				},
			},
			"write_to_logs_streams": schema.BoolAttribute{
				Description: "When type is remote_elasticsearch, whether agents using this output send data to wired logs streams (preview in newer stacks).",
				Optional:    true,
				Validators: []validator.Bool{
					validators.AllowedIfDependentPathEquals(path.Root("type"), "remote_elasticsearch", validators.AllowedIfOptions{}),
				},
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
				Description: "Advanced YAML configuration. YAML settings here will be added to the output section of each agent policy. " +
					"Note: the Fleet API treats an omitted `config_yaml` in an update request as \"no change\" and does not clear the stored value. " +
					"To clear a previously stored value, delete and re-create the output.",
				Optional:   true,
				Sensitive:  true,
				CustomType: customtypes.NormalizedYamlType{},
			},
			"space_ids": schema.SetAttribute{
				Description: spaceIDsDescription,
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			attrSSL: schema.SingleNestedAttribute{
				Description: "SSL configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					attrCertificateAuthorities: schema.ListAttribute{
						Description: "Server SSL certificate authorities.",
						Optional:    true,
						ElementType: types.StringType,
						Validators: []validator.List{
							listvalidator.SizeAtLeast(1),
						},
					},
					attrCertificate: schema.StringAttribute{
						Description: "Client SSL certificate.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					attrKey: schema.StringAttribute{
						Description: "Client SSL certificate key.",
						Optional:    true,
						Sensitive:   true,
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
					attrVerificationMode: schema.StringAttribute{
						Description: "The SSL verification mode. One of `certificate`, `full`, `none`, `strict`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("certificate", "full", "none", "strict"),
						},
					},
				},
			},
			attrKafka: schema.SingleNestedAttribute{
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
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
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
							kafkaCompressionLevelDefaultIfGzip(),
						},
						Validators: []validator.Int64{
							validators.AllowedIfDependentPathEquals(path.Root("kafka").AtName("compression"), "gzip", validators.AllowedIfOptions{}),
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
								validators.AllowedIfOptions{},
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
						Computed:    true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("random", "round_robin", "hash"),
						},
					},
					"required_acks": schema.Int64Attribute{
						Description: "Number of acknowledgments required for Kafka output.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
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
					attrKey: schema.StringAttribute{
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
								attrKey: schema.StringAttribute{
									Description: "Header key.",
									Required:    true,
								},
								attrValue: schema.StringAttribute{
									Description: "Header value.",
									Required:    true,
								},
							},
						},
					},
					attrHash: schema.SingleNestedAttribute{
						Description: "Hash configuration for Kafka partition.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							attrHash: schema.StringAttribute{
								Description: "Hash field.",
								Optional:    true,
							},
							attrRandom: schema.BoolAttribute{
								Description: "Use random hash.",
								Optional:    true,
							},
						},
					},
					attrRandom: schema.SingleNestedAttribute{
						Description: "Random configuration for Kafka partition.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							attrGroupEvents: schema.Float64Attribute{
								Description: "Number of events to group.",
								Optional:    true,
							},
						},
					},
					"round_robin": schema.SingleNestedAttribute{
						Description: "Round robin configuration for Kafka partition.",
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							attrGroupEvents: schema.Float64Attribute{
								Description: "Number of events to group.",
								Optional:    true,
							},
						},
					},
					"sasl": schema.SingleNestedAttribute{
						Description: "SASL configuration for Kafka authentication.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							attrMechanism: schema.StringAttribute{
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

func getSslAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["ssl"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getHeadersAttrTypes(ctx context.Context) attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["headers"].GetType().(attr.TypeWithElementType).ElementType()
}

func getHashAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["hash"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getRandomAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["random"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getRoundRobinAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["round_robin"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getSaslAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).Attributes["sasl"].GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}

func getKafkaAttrTypes(ctx context.Context) map[string]attr.Type {
	return getSchema(ctx).Attributes["kafka"].(schema.SingleNestedAttribute).GetType().(attr.TypeWithAttributeTypes).AttributeTypes()
}
