package config

import (
	"context"
	"os"
	"testing"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func Test_newKibanaConfigFromSDK(t *testing.T) {
	type args struct {
		baseCfg        baseConfig
		resourceData   map[string]interface{}
		expectedConfig kibanaConfig
		expectedDiags  sdkdiags.Diagnostics
		env            map[string]string
	}
	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "should return kibana config if no fleet config defined",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg:        baseCfg,
					resourceData:   map[string]interface{}{},
					expectedConfig: baseCfg.toKibanaConfig(),
				}
			},
		},
		{
			name: "should use the provided config options",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"username":  "kibana",
								"password":  "baltic",
								"ca_certs":  []interface{}{"internal", "lets_decrypt"},
								"insecure":  false,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana",
						Password:         "baltic",
						CAs:              []string{"internal", "lets_decrypt"},
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should prefer environment variables",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"username":  "kibana",
								"password":  "baltic",
								"ca_certs":  []interface{}{"internal", "lets_decrypt"},
								"insecure":  true,
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT": "example.com/cabana",
						"KIBANA_USERNAME": "elastic",
						"KIBANA_PASSWORD": "thin-lines",
						"KIBANA_INSECURE": "false",
						"KIBANA_CA_CERTS": "black,sea",
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/cabana",
						Username:         "elastic",
						Password:         "thin-lines",
						DisableVerifySSL: false,
						CAs:              []string{"black", "sea"},
					},
				}
			},
		},
		{
			name: "should fallback to elasticsearch username/password when no kibana credentials provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								// No username/password provided in kibana config
								"ca_certs": []interface{}{"internal"},
								"insecure": false,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "es-user",     // Falls back to ES credentials
						Password:         "es-password", // Falls back to ES credentials
						CAs:              []string{"internal"},
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should fallback to elasticsearch api_key when no kibana credentials provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								// No api_key provided in kibana config
								"insecure": true,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "es-api-key-123", // Falls back to ES api_key
						DisableVerifySSL: true,
					},
				}
			},
		},
		{
			name: "should not override kibana credentials when they are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"username":  "kibana-user",
								"password":  "kibana-password",
								"insecure":  false,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana-user",     // Uses kibana-specific credentials
						Password:         "kibana-password", // Uses kibana-specific credentials
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not override kibana api_key when explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"api_key":   "es-api-key-456",
								"insecure":  false,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "es-api-key-456", // Uses kibana-specific api_key
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not add elasticsearch api key when kibana credentials are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"username":  "kibana-user",
								"password":  "kibana-password",
								"insecure":  false,
							},
						},
					},
					env: map[string]string{},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana-user",     // Uses kibana-specific credentials
						Password:         "kibana-password", // Uses kibana-specific credentials
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not add elasticsearch credentials when kibana api key are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"insecure":  false,
								"api_key":   "es-api-key-123",
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "es-api-key-123",
						DisableVerifySSL: false,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("KIBANA_USERNAME")
			os.Unsetenv("KIBANA_PASSWORD")
			os.Unsetenv("KIBANA_ENDPOINT")
			os.Unsetenv("KIBANA_INSECURE")
			os.Unsetenv("KIBANA_API_KEY")
			os.Unsetenv("KIBANA_CA_CERTS")

			args := tt.args()
			rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"kibana": providerSchema.GetKibanaConnectionSchema(),
			}, args.resourceData)

			for key, val := range args.env {
				os.Setenv(key, val)
			}

			kibanaCfg, diags := newKibanaConfigFromSDK(rd, args.baseCfg)

			require.Equal(t, args.expectedConfig, kibanaCfg)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newKibanaConfigFromFramework(t *testing.T) {
	type args struct {
		baseCfg        baseConfig
		providerConfig ProviderConfiguration
		expectedConfig kibanaConfig
		expectedDiags  fwdiags.Diagnostics
		env            map[string]string
	}
	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "should return kibana config if no fleet config defined",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg:        baseCfg,
					providerConfig: ProviderConfiguration{},
					expectedConfig: baseCfg.toKibanaConfig(),
				}
			},
		},
		{
			name: "should use the provided config options",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Username: types.StringValue("kibana"),
								Password: types.StringValue("baltic"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
									types.StringValue("lets_decrypt"),
								}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana",
						Password:         "baltic",
						CAs:              []string{"internal", "lets_decrypt"},
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should use api_key when provided in config options",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "test",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								ApiKey: types.StringValue("test"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(true),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "test",
						DisableVerifySSL: true,
					},
				}
			},
		},
		{
			name: "should prefer environment variables",
			args: func() args {
				baseCfg := baseConfig{
					Username: "elastic",
					Password: "changeme",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Username: types.StringValue("kibana"),
								Password: types.StringValue("baltic"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
									types.StringValue("lets_decrypt"),
								}),
								Insecure: types.BoolValue(true),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT": "example.com/cabana",
						"KIBANA_USERNAME": "elastic",
						"KIBANA_PASSWORD": "thin-lines",
						"KIBANA_INSECURE": "false",
						"KIBANA_CA_CERTS": "black,sea",
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/cabana",
						Username:         "elastic",
						Password:         "thin-lines",
						CAs:              []string{"black", "sea"},
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should fallback to elasticsearch username/password when no kibana credentials provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								// No username/password provided in kibana config
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
								}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "es-user",     // Falls back to ES credentials
						Password:         "es-password", // Falls back to ES credentials
						CAs:              []string{"internal"},
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should fallback to elasticsearch api_key when no kibana credentials provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								// No api_key provided in kibana config
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(true),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "es-api-key-123", // Falls back to ES api_key
						DisableVerifySSL: true,
					},
				}
			},
		},
		{
			name: "should not override kibana credentials when they are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Username: types.StringValue("kibana-user"),
								Password: types.StringValue("kibana-password"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana-user",     // Uses kibana-specific credentials
						Password:         "kibana-password", // Uses kibana-specific credentials
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not override kibana api_key when explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								ApiKey: types.StringValue("kibana-api-key-456"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "kibana-api-key-456", // Uses kibana-specific api_key
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not add elasticsearch api key when kibana credentials are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					ApiKey: "es-api-key-123",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Username: types.StringValue("kibana-user"),
								Password: types.StringValue("kibana-password"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					env: map[string]string{},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana-user",     // Uses kibana-specific credentials
						Password:         "kibana-password", // Uses kibana-specific credentials
						DisableVerifySSL: false,
					},
				}
			},
		},
		{
			name: "should not add elasticsearch credentials when kibana api key are explicitly provided",
			args: func() args {
				baseCfg := baseConfig{
					Username: "es-user",
					Password: "es-password",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								ApiKey: types.StringValue("kibana-api-key-456"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						ApiKey:           "kibana-api-key-456",
						DisableVerifySSL: false,
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("KIBANA_USERNAME")
			os.Unsetenv("KIBANA_PASSWORD")
			os.Unsetenv("KIBANA_API_KEY")
			os.Unsetenv("KIBANA_ENDPOINT")
			os.Unsetenv("KIBANA_CA_CERTS")
			os.Unsetenv("KIBANA_INSECURE")

			args := tt.args()

			for key, val := range args.env {
				os.Setenv(key, val)
			}

			kibanaCfg, diags := newKibanaConfigFromFramework(context.Background(), args.providerConfig, args.baseCfg)

			require.Equal(t, args.expectedConfig, kibanaCfg)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}
