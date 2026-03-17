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
		resourceData   map[string]any
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
					resourceData:   map[string]any{},
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
					resourceData: map[string]any{
						"kibana": []any{
							map[string]any{
								"endpoints": []any{"example.com/kibana"},
								"username":  "kibana",
								"password":  "baltic",
								"ca_certs":  []any{"internal", "lets_decrypt"},
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
					resourceData: map[string]any{
						"kibana": []any{
							map[string]any{
								"endpoints": []any{"example.com/kibana"},
								"username":  "kibana",
								"password":  "baltic",
								"ca_certs":  []any{"internal", "lets_decrypt"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("KIBANA_USERNAME")
			os.Unsetenv("KIBANA_PASSWORD")
			os.Unsetenv("KIBANA_ENDPOINT")
			os.Unsetenv("KIBANA_INSECURE")
			os.Unsetenv("KIBANA_API_KEY")
			os.Unsetenv("KIBANA_BEARER_TOKEN")
			os.Unsetenv("KIBANA_CA_CERTS")

			args := tt.args()
			rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"kibana": providerSchema.GetKibanaConnectionSchema(),
			}, args.resourceData)

			for key, val := range args.env {
				t.Setenv(key, val)
			}

			kibanaCfg, diags := newKibanaConfigFromSDK(rd, args.baseCfg)

			require.Equal(t, args.expectedConfig, kibanaCfg)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newKibanaConfigFromSDK_BearerToken(t *testing.T) {
	os.Unsetenv("KIBANA_USERNAME")
	os.Unsetenv("KIBANA_PASSWORD")
	os.Unsetenv("KIBANA_ENDPOINT")
	os.Unsetenv("KIBANA_INSECURE")
	os.Unsetenv("KIBANA_API_KEY")
	os.Unsetenv("KIBANA_BEARER_TOKEN")
	os.Unsetenv("KIBANA_CA_CERTS")

	baseCfg := baseConfig{}
	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"kibana": providerSchema.GetKibanaConnectionSchema(),
	}, map[string]any{
		"kibana": []any{
			map[string]any{
				"endpoints":    []any{"example.com/kibana"},
				"bearer_token": "my-jwt-token",
				"insecure":     true,
			},
		},
	})

	kibanaCfg, diags := newKibanaConfigFromSDK(rd, baseCfg)

	require.Nil(t, diags)
	require.Equal(t, "my-jwt-token", kibanaCfg.BearerToken)
	require.Equal(t, "example.com/kibana", kibanaCfg.Address)
	require.True(t, kibanaCfg.DisableVerifySSL)
}

func Test_newKibanaConfigFromSDK_BearerTokenEnvOverride(t *testing.T) {
	os.Unsetenv("KIBANA_USERNAME")
	os.Unsetenv("KIBANA_PASSWORD")
	os.Unsetenv("KIBANA_ENDPOINT")
	os.Unsetenv("KIBANA_INSECURE")
	os.Unsetenv("KIBANA_API_KEY")
	os.Unsetenv("KIBANA_BEARER_TOKEN")
	os.Unsetenv("KIBANA_CA_CERTS")

	t.Setenv("KIBANA_BEARER_TOKEN", "env-jwt-token")

	baseCfg := baseConfig{}
	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"kibana": providerSchema.GetKibanaConnectionSchema(),
	}, map[string]any{
		"kibana": []any{
			map[string]any{
				"endpoints":    []any{"example.com/kibana"},
				"bearer_token": "config-jwt-token",
			},
		},
	})

	kibanaCfg, diags := newKibanaConfigFromSDK(rd, baseCfg)

	require.Nil(t, diags)
	require.Equal(t, "env-jwt-token", kibanaCfg.BearerToken)
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
					APIKey: "test",
				}

				return args{
					baseCfg: baseCfg,
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								APIKey: types.StringValue("test"),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("KIBANA_USERNAME")
			os.Unsetenv("KIBANA_PASSWORD")
			os.Unsetenv("KIBANA_API_KEY")
			os.Unsetenv("KIBANA_BEARER_TOKEN")
			os.Unsetenv("KIBANA_ENDPOINT")
			os.Unsetenv("KIBANA_CA_CERTS")
			os.Unsetenv("KIBANA_INSECURE")

			args := tt.args()

			for key, val := range args.env {
				t.Setenv(key, val)
			}

			kibanaCfg, diags := newKibanaConfigFromFramework(context.Background(), args.providerConfig, args.baseCfg)

			require.Equal(t, args.expectedConfig, kibanaCfg)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newKibanaConfigFromFramework_BearerToken(t *testing.T) {
	os.Unsetenv("KIBANA_USERNAME")
	os.Unsetenv("KIBANA_PASSWORD")
	os.Unsetenv("KIBANA_API_KEY")
	os.Unsetenv("KIBANA_BEARER_TOKEN")
	os.Unsetenv("KIBANA_ENDPOINT")
	os.Unsetenv("KIBANA_CA_CERTS")
	os.Unsetenv("KIBANA_INSECURE")

	baseCfg := baseConfig{}
	providerConfig := ProviderConfiguration{
		Kibana: []KibanaConnection{
			{
				BearerToken: types.StringValue("my-jwt-token"),
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("example.com/kibana"),
				}),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure: types.BoolValue(true),
			},
		},
	}

	kibanaCfg, diags := newKibanaConfigFromFramework(context.Background(), providerConfig, baseCfg)

	require.False(t, diags.HasError())
	require.Equal(t, "my-jwt-token", kibanaCfg.BearerToken)
	require.Equal(t, "example.com/kibana", kibanaCfg.Address)
	require.True(t, kibanaCfg.DisableVerifySSL)
}

func Test_newKibanaConfigFromFramework_BearerTokenEnvOverride(t *testing.T) {
	os.Unsetenv("KIBANA_USERNAME")
	os.Unsetenv("KIBANA_PASSWORD")
	os.Unsetenv("KIBANA_API_KEY")
	os.Unsetenv("KIBANA_BEARER_TOKEN")
	os.Unsetenv("KIBANA_ENDPOINT")
	os.Unsetenv("KIBANA_CA_CERTS")
	os.Unsetenv("KIBANA_INSECURE")

	t.Setenv("KIBANA_BEARER_TOKEN", "env-jwt-token")

	baseCfg := baseConfig{}
	providerConfig := ProviderConfiguration{
		Kibana: []KibanaConnection{
			{
				BearerToken: types.StringValue("config-jwt-token"),
				Endpoints: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("example.com/kibana"),
				}),
				CACerts: types.ListValueMust(types.StringType, []attr.Value{}),
			},
		},
	}

	kibanaCfg, diags := newKibanaConfigFromFramework(context.Background(), providerConfig, baseCfg)

	require.False(t, diags.HasError())
	require.Equal(t, "env-jwt-token", kibanaCfg.BearerToken)
}
