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
			name: "should use the provided config optios",
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
								"insecure":  true,
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana",
						Password:         "baltic",
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
					resourceData: map[string]interface{}{
						"kibana": []interface{}{
							map[string]interface{}{
								"endpoints": []interface{}{"example.com/kibana"},
								"username":  "kibana",
								"password":  "baltic",
								"insecure":  true,
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT": "example.com/cabana",
						"KIBANA_USERNAME": "elastic",
						"KIBANA_PASSWORD": "thin-lines",
						"KIBANA_INSECURE": "false",
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/cabana",
						Username:         "elastic",
						Password:         "thin-lines",
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
								Insecure: types.BoolValue(true),
							},
						},
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/kibana",
						Username:         "kibana",
						Password:         "baltic",
						DisableVerifySSL: true,
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
								Insecure: types.BoolValue(true),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT": "example.com/cabana",
						"KIBANA_USERNAME": "elastic",
						"KIBANA_PASSWORD": "thin-lines",
						"KIBANA_INSECURE": "false",
					},
					expectedConfig: kibanaConfig{
						Address:          "example.com/cabana",
						Username:         "elastic",
						Password:         "thin-lines",
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
