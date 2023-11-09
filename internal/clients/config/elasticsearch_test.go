package config

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/stretchr/testify/require"
)

func Test_newElasticsearchConfigFromSDK(t *testing.T) {
	type args struct {
		resourceData     map[string]interface{}
		base             baseConfig
		env              map[string]string
		expectedESConfig *elasticsearchConfig
		expectedDiags    sdkdiags.Diagnostics
	}
	tests := []struct {
		name string
		args func(string) args
	}{
		{
			name: "should return nil if no config is specified",
			args: func(key string) args {
				return args{}
			},
		},
		{
			name: "should use the options set in config",
			args: func(key string) args {
				base := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				config := base.toElasticsearchConfig()
				config.Addresses = []string{"localhost", "example.com"}
				tlsConfig := config.ensureTLSClientConfig()
				tlsConfig.InsecureSkipVerify = true

				return args{
					resourceData: map[string]interface{}{
						key: []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"localhost", "example.com"},
								"insecure":  true,
							},
						},
					},
					base:             base,
					expectedESConfig: &config,
				}
			},
		},
		{
			name: "should prefer config defined in environment variables",
			args: func(key string) args {
				base := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				config := base.toElasticsearchConfig()
				config.Addresses = []string{"127.0.0.1", "example.com/elastic"}
				tlsConfig := config.ensureTLSClientConfig()
				tlsConfig.InsecureSkipVerify = false

				return args{
					resourceData: map[string]interface{}{
						key: []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"localhost", "example.com"},
								"insecure":  true,
							},
						},
					},
					env: map[string]string{
						"ELASTICSEARCH_ENDPOINTS": "127.0.0.1,example.com/elastic",
						"ELASTICSEARCH_INSECURE":  "false",
					},
					base:             base,
					expectedESConfig: &config,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("ELASTICSEARCH_ENDPOINTS")
			os.Unsetenv("ELASTICSEARCH_INSECURE")

			key := "elasticsearch"
			args := tt.args(key)
			rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				key: providerSchema.GetEsConnectionSchema(key, true),
			}, args.resourceData)

			for key, val := range args.env {
				os.Setenv(key, val)
			}

			esConfig, diags := newElasticsearchConfigFromSDK(rd, args.base, key, false)

			require.Equal(t, args.expectedESConfig, esConfig)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newElasticsearchConfigFromFramework(t *testing.T) {
	type args struct {
		providerConfig   ProviderConfiguration
		base             baseConfig
		env              map[string]string
		expectedESConfig *elasticsearchConfig
		expectedDiags    fwdiags.Diagnostics
	}
	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "should return nil if no config is specified",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{},
				}
			},
		},
		{
			name: "should use the options set in config",
			args: func() args {
				base := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				config := base.toElasticsearchConfig()
				config.Addresses = []string{"localhost", "example.com"}
				tlsConfig := config.ensureTLSClientConfig()
				tlsConfig.InsecureSkipVerify = true

				return args{
					providerConfig: ProviderConfiguration{
						Elasticsearch: []ElasticsearchConnection{
							{
								Endpoints: basetypes.NewListValueMust(
									basetypes.StringType{},
									[]attr.Value{
										basetypes.NewStringValue("localhost"),
										basetypes.NewStringValue("example.com"),
									},
								),
								Insecure: basetypes.NewBoolPointerValue(utils.Pointer(true)),
							},
						},
					},
					base:             base,
					expectedESConfig: &config,
				}
			},
		},
		{
			name: "should prefer config defined in environment variables",
			args: func() args {
				base := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				config := base.toElasticsearchConfig()
				config.Addresses = []string{"127.0.0.1", "example.com/elastic"}
				tlsConfig := config.ensureTLSClientConfig()
				tlsConfig.InsecureSkipVerify = false

				return args{
					providerConfig: ProviderConfiguration{
						Elasticsearch: []ElasticsearchConnection{
							{
								Endpoints: basetypes.NewListValueMust(
									basetypes.StringType{},
									[]attr.Value{
										basetypes.NewStringValue("localhost"),
										basetypes.NewStringValue("example.com"),
									},
								),
								Insecure: basetypes.NewBoolPointerValue(utils.Pointer(true)),
							},
						},
					},
					env: map[string]string{
						"ELASTICSEARCH_ENDPOINTS": "127.0.0.1,example.com/elastic",
						"ELASTICSEARCH_INSECURE":  "false",
					},
					base:             base,
					expectedESConfig: &config,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("ELASTICSEARCH_ENDPOINTS")
			os.Unsetenv("ELASTICSEARCH_INSECURE")

			args := tt.args()

			for key, val := range args.env {
				os.Setenv(key, val)
			}

			esConfig, diags := newElasticsearchConfigFromFramework(context.Background(), args.providerConfig, args.base)

			require.Equal(t, args.expectedESConfig, esConfig)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}
