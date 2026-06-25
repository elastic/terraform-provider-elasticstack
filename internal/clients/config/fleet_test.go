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

func fleetMultipleAuthWarningDiag() fwdiags.Diagnostics {
	var d fwdiags.Diagnostics
	addMultipleAuthWarning(&d, "Fleet", "Fleet environment variables")
	return d
}

func Test_newFleetConfigFromFramework(t *testing.T) {
	type args struct {
		kibanaCfg      kibanaOapiConfig
		providerConfig ProviderConfiguration
		expectedConfig fleetConfig
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
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "elastic",
					Password: "changeme",
					Insecure: true,
				}

				return args{
					kibanaCfg:      kibanaCfg,
					providerConfig: ProviderConfiguration{},
					expectedConfig: kibanaCfg.toFleetConfig(),
				}
			},
		},
		{
			name: "should use the provided config options",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "elastic",
					Password: "changeme",
					Insecure: true,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet"),
								Password: types.StringValue("baltic"),
								Endpoint: types.StringValue("example.com/fleet"),
								APIKey:   types.StringValue("leviosa"),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
									types.StringValue("lets_decrypt"),
								}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "fleet",
						Password: "baltic",
						APIKey:   "leviosa",
						CACerts:  []string{"internal", "lets_decrypt"},
						Insecure: false,
					},
					expectedDiags: fleetMultipleAuthWarningDiag(),
				}
			},
		},
		{
			name: "should prefer environment variables",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "elastic",
					Password: "changeme",
					Insecure: true,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet"),
								Password: types.StringValue("baltic"),
								Endpoint: types.StringValue("example.com/fleet"),
								APIKey:   types.StringValue("leviosa"),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
									types.StringValue("lets_decrypt"),
								}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					env: map[string]string{
						"FLEET_ENDPOINT": "example.com/black_sea_fleet",
						"FLEET_USERNAME": "black_sea",
						"FLEET_PASSWORD": "fleet",
						"FLEET_API_KEY":  "stupefy",
						"FLEET_CA_CERTS": "black,sea",
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/black_sea_fleet",
						Username: "black_sea",
						Password: "fleet",
						APIKey:   "stupefy",
						CACerts:  []string{"black", "sea"},
						Insecure: false,
					},
					expectedDiags: fleetMultipleAuthWarningDiag(),
				}
			},
		},
		{
			name: "Kibana BasicAuth + Fleet APIKey clears inherited BasicAuth",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					Username: "k",
					Password: "p",
				}
				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								APIKey:  types.StringValue("fleet-key"),
								CACerts: types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					expectedConfig: fleetConfig{
						APIKey:   "fleet-key",
						Username: "",
						Password: "",
					},
				}
			},
		},
		{
			name: "Fleet env password + provider username overrides inherited Kibana APIKey",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{APIKey: "kibana-key"}
				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"FLEET_PASSWORD": "env-pass",
					},
					expectedConfig: fleetConfig{
						Username: "fleet-user",
						Password: "env-pass",
						APIKey:   "",
					},
				}
			},
		},
		{
			name: "FLEET_API_KEY env overrides provider Fleet BasicAuth",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{}
				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								Password: types.StringValue("fleet-pass"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"FLEET_API_KEY": "env-key",
					},
					expectedConfig: fleetConfig{
						APIKey:   "env-key",
						Username: "",
						Password: "",
					},
				}
			},
		},
		{
			name: "env-level conflict FLEET_API_KEY + FLEET_USERNAME emits warning",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{}
				return args{
					kibanaCfg:      kibanaCfg,
					providerConfig: ProviderConfiguration{},
					env: map[string]string{
						"FLEET_API_KEY":  "env-key",
						"FLEET_USERNAME": "env-user",
					},
					expectedConfig: fleetConfig{
						APIKey:   "env-key",
						Username: "env-user",
					},
					expectedDiags: fleetMultipleAuthWarningDiag(),
				}
			},
		},
		{
			name: "inherited Kibana config with multiple auth methods emits Fleet warning",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					APIKey:   "bad-key",
					Username: "bad-user",
				}
				return args{
					kibanaCfg:      kibanaCfg,
					providerConfig: ProviderConfiguration{},
					expectedConfig: fleetConfig{
						APIKey:   "bad-key",
						Username: "bad-user",
					},
					expectedDiags: fleetMultipleAuthWarningDiag(),
				}
			},
		},
		{
			name: "FLEET_BEARER_TOKEN env overrides provider Fleet BasicAuth",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{}
				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								Password: types.StringValue("fleet-pass"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
							},
						},
					},
					env: map[string]string{
						"FLEET_BEARER_TOKEN": "env-token",
					},
					expectedConfig: fleetConfig{
						BearerToken: "env-token",
						Username:    "",
						Password:    "",
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("FLEET_ENDPOINT")
			os.Unsetenv("FLEET_USERNAME")
			os.Unsetenv("FLEET_PASSWORD")
			os.Unsetenv("FLEET_API_KEY")
			os.Unsetenv("FLEET_BEARER_TOKEN")
			os.Unsetenv("FLEET_CA_CERTS")

			args := tt.args()

			for key, val := range args.env {
				t.Setenv(key, val)
			}

			fleetConfig, diags := newFleetConfigFromFramework(context.Background(), args.providerConfig, args.kibanaCfg)

			require.Equal(t, args.expectedConfig, fleetConfig)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newFleetConfigFromFramework_BearerToken(t *testing.T) {
	os.Unsetenv("FLEET_ENDPOINT")
	os.Unsetenv("FLEET_USERNAME")
	os.Unsetenv("FLEET_PASSWORD")
	os.Unsetenv("FLEET_API_KEY")
	os.Unsetenv("FLEET_BEARER_TOKEN")
	os.Unsetenv("FLEET_CA_CERTS")

	kibanaCfg := kibanaOapiConfig{}
	providerConfig := ProviderConfiguration{
		Fleet: []FleetConnection{
			{
				BearerToken: types.StringValue("my-jwt-token"),
				Endpoint:    types.StringValue("example.com/fleet"),
				CACerts:     types.ListValueMust(types.StringType, []attr.Value{}),
				Insecure:    types.BoolValue(true),
			},
		},
	}

	fleetCfg, diags := newFleetConfigFromFramework(context.Background(), providerConfig, kibanaCfg)

	require.False(t, diags.HasError())
	require.Equal(t, "my-jwt-token", fleetCfg.BearerToken)
	require.Equal(t, "example.com/fleet", fleetCfg.URL)
	require.True(t, fleetCfg.Insecure)
}

func Test_newFleetConfigFromFramework_BearerTokenEnvOverride(t *testing.T) {
	os.Unsetenv("FLEET_ENDPOINT")
	os.Unsetenv("FLEET_USERNAME")
	os.Unsetenv("FLEET_PASSWORD")
	os.Unsetenv("FLEET_API_KEY")
	os.Unsetenv("FLEET_BEARER_TOKEN")
	os.Unsetenv("FLEET_CA_CERTS")

	t.Setenv("FLEET_BEARER_TOKEN", "env-jwt-token")

	kibanaCfg := kibanaOapiConfig{}
	providerConfig := ProviderConfiguration{
		Fleet: []FleetConnection{
			{
				BearerToken: types.StringValue("config-jwt-token"),
				Endpoint:    types.StringValue("example.com/fleet"),
				CACerts:     types.ListValueMust(types.StringType, []attr.Value{}),
			},
		},
	}

	fleetCfg, diags := newFleetConfigFromFramework(context.Background(), providerConfig, kibanaCfg)

	require.False(t, diags.HasError())
	require.Equal(t, "env-jwt-token", fleetCfg.BearerToken)
}
