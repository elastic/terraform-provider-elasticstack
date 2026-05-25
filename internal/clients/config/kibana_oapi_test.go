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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func Test_newKibanaOapiConfigFromFramework(t *testing.T) {
	type args struct {
		baseCfg        baseConfig
		providerConfig ProviderConfiguration
		expectedConfig kibanaOapiConfig
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
					expectedConfig: baseCfg.toKibanaOapiConfig(),
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
					expectedConfig: kibanaOapiConfig{
						URL:      "example.com/kibana",
						Username: "kibana",
						Password: "baltic",
						CACerts:  []string{"internal", "lets_decrypt"},
						Insecure: false,
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
					expectedConfig: kibanaOapiConfig{
						URL:      "example.com/kibana",
						APIKey:   "test",
						Insecure: true,
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
					expectedConfig: kibanaOapiConfig{
						URL:      "example.com/cabana",
						Username: "elastic",
						Password: "thin-lines",
						CACerts:  []string{"black", "sea"},
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should keep configured endpoint when explicitly requested",
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
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("example.com/kibana"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT":                    "example.com/cabana",
						PreferConfiguredKibanaEndpointEnvVar: "true",
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "example.com/kibana",
						Username: "elastic",
						Password: "changeme",
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
			os.Unsetenv(PreferConfiguredKibanaEndpointEnvVar)

			args := tt.args()

			for key, val := range args.env {
				t.Setenv(key, val)
			}

			kibanaCfg, diags := newKibanaOapiConfigFromFramework(context.Background(), args.providerConfig, args.baseCfg)

			require.Equal(t, args.expectedConfig, kibanaCfg)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newKibanaOapiConfigFromFramework_doesNotApplyFleetFallback(t *testing.T) {
	os.Unsetenv("KIBANA_ENDPOINT")

	cfg := ProviderConfiguration{
		Fleet: []FleetConnection{
			{
				Endpoint: types.StringValue("https://fleet.example.com"),
				APIKey:   types.StringValue("F-KEY"),
				CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
			},
		},
	}

	kibanaCfg, diags := newKibanaOapiConfigFromFramework(context.Background(), cfg, baseConfig{})

	require.False(t, diags.HasError())
	require.Empty(t, kibanaCfg.URL)
	require.Empty(t, kibanaCfg.APIKey)
}

func Test_newProviderKibanaOapiConfigFromFramework_fleetBlockFallback(t *testing.T) {
	type args struct {
		baseCfg        baseConfig
		providerConfig ProviderConfiguration
		expectedConfig kibanaOapiConfig
		env            map[string]string
	}

	tests := []struct {
		name string
		args func() args
	}{
		{
			name: "fleet endpoint only inherits into kibana_oapi URL",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL: "https://fleet.example.com",
					},
				}
			},
		},
		{
			name: "kibana endpoints with fleet api_key uses both blocks field-by-field",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("https://kibana.example.com"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
						Fleet: []FleetConnection{
							{
								APIKey:  types.StringValue("F-KEY"),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://kibana.example.com",
						APIKey:   "F-KEY",
						Insecure: false,
					},
				}
			},
		},
		{
			name: "KIBANA_ENDPOINT env override wins over fleet-block URL fallback",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT": "https://env.example.com",
					},
					expectedConfig: kibanaOapiConfig{
						URL: "https://env.example.com",
					},
				}
			},
		},
		{
			name: "prefer configured kibana endpoint keeps kibana URL over env when both kibana and fleet blocks set URL sources",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("https://kibana.example.com"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT":                    "https://env.example.com",
						PreferConfiguredKibanaEndpointEnvVar: "true",
					},
					expectedConfig: kibanaOapiConfig{
						URL: "https://kibana.example.com",
					},
				}
			},
		},
		{
			name: "fleet-only URL with prefer configured and KIBANA_ENDPOINT env uses env URL",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"KIBANA_ENDPOINT":                    "https://env.example.com",
						PreferConfiguredKibanaEndpointEnvVar: "true",
					},
					expectedConfig: kibanaOapiConfig{
						URL: "https://env.example.com",
					},
				}
			},
		},
		{
			name: "kibana insecure unset inherits fleet insecure true",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("https://kibana.example.com"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolNull(),
							},
						},
						Fleet: []FleetConnection{
							{
								Insecure: types.BoolValue(true),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://kibana.example.com",
						Insecure: true,
					},
				}
			},
		},
		{
			name: "kibana insecure false is not overridden by fleet insecure true",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("https://kibana.example.com"),
								}),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
						Fleet: []FleetConnection{
							{
								Insecure: types.BoolValue(true),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://kibana.example.com",
						Insecure: false,
					},
				}
			},
		},
		{
			name: "fleet-only insecure true inherits into kibana_oapi",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								Insecure: types.BoolValue(true),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://fleet.example.com",
						Insecure: true,
					},
				}
			},
		},
		{
			name: "fleet-only username and password inherit into kibana_oapi",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								Password: types.StringValue("fleet-pass"),
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://fleet.example.com",
						Username: "fleet-user",
						Password: "fleet-pass",
					},
				}
			},
		},
		{
			name: "fleet-only bearer token inherits into kibana_oapi",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								BearerToken: types.StringValue("fleet-jwt"),
								Endpoint:    types.StringValue("https://fleet.example.com"),
								CACerts:     types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:         "https://fleet.example.com",
						BearerToken: "fleet-jwt",
					},
				}
			},
		},
		{
			name: "fleet-only ca_certs inherit into kibana_oapi",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("https://fleet.example.com"),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("fleet-ca-1"),
									types.StringValue("fleet-ca-2"),
								}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:     "https://fleet.example.com",
						CACerts: []string{"fleet-ca-1", "fleet-ca-2"},
					},
				}
			},
		},
		{
			name: "kibana username with fleet password uses both blocks field-by-field",
			args: func() args {
				return args{
					providerConfig: ProviderConfiguration{
						Kibana: []KibanaConnection{
							{
								Username: types.StringValue("kibana-user"),
								Endpoints: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("https://kibana.example.com"),
								}),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
						Fleet: []FleetConnection{
							{
								Password: types.StringValue("fleet-pass"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: kibanaOapiConfig{
						URL:      "https://kibana.example.com",
						Username: "kibana-user",
						Password: "fleet-pass",
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
			os.Unsetenv(PreferConfiguredKibanaEndpointEnvVar)

			args := tt.args()

			for key, val := range args.env {
				t.Setenv(key, val)
			}

			kibanaCfg, diags := newProviderKibanaOapiConfigFromFramework(context.Background(), args.providerConfig, args.baseCfg)

			require.False(t, diags.HasError())
			require.Equal(t, args.expectedConfig, kibanaCfg)
		})
	}
}
