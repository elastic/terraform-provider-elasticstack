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

func Test_newFleetConfigFromSDK(t *testing.T) {
	type args struct {
		kibanaCfg      kibanaOapiConfig
		resourceData   map[string]interface{}
		expectedConfig fleetConfig
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
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "elastic",
					Password: "changeme",
					Insecure: true,
				}

				return args{
					kibanaCfg:      kibanaCfg,
					resourceData:   map[string]interface{}{},
					expectedConfig: kibanaCfg.toFleetConfig(),
				}
			},
		},
		{
			name: "should use the provided config optios",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "elastic",
					Password: "changeme",
					Insecure: true,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"username": "fleet",
								"password": "baltic",
								"api_key":  "leviosa",
								"ca_certs": []interface{}{"internal", "lets_decrypt"},
								"insecure": false,
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
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"username": "fleet",
								"password": "baltic",
								"api_key":  "leviosa",
								"ca_certs": []interface{}{"internal", "lets_decrypt"},
								"insecure": false,
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
			rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
				"fleet": providerSchema.GetFleetConnectionSchema(),
			}, args.resourceData)

			for key, val := range args.env {
				os.Setenv(key, val)
			}

			fleetConfig, diags := newFleetConfigFromSDK(rd, args.kibanaCfg)

			require.Equal(t, args.expectedConfig, fleetConfig)
			require.Equal(t, args.expectedDiags, diags)
		})
	}
}

func Test_newFleetConfigFromSDK_BearerToken(t *testing.T) {
	os.Unsetenv("FLEET_ENDPOINT")
	os.Unsetenv("FLEET_USERNAME")
	os.Unsetenv("FLEET_PASSWORD")
	os.Unsetenv("FLEET_API_KEY")
	os.Unsetenv("FLEET_BEARER_TOKEN")
	os.Unsetenv("FLEET_CA_CERTS")

	kibanaCfg := kibanaOapiConfig{}
	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"fleet": providerSchema.GetFleetConnectionSchema(),
	}, map[string]interface{}{
		"fleet": []interface{}{
			map[string]interface{}{
				"endpoint":     "example.com/fleet",
				"bearer_token": "my-jwt-token",
				"insecure":     true,
			},
		},
	})

	fleetCfg, diags := newFleetConfigFromSDK(rd, kibanaCfg)

	require.Nil(t, diags)
	require.Equal(t, "my-jwt-token", fleetCfg.BearerToken)
	require.Equal(t, "example.com/fleet", fleetCfg.URL)
	require.True(t, fleetCfg.Insecure)
}

func Test_newFleetConfigFromSDK_BearerTokenEnvOverride(t *testing.T) {
	os.Unsetenv("FLEET_ENDPOINT")
	os.Unsetenv("FLEET_USERNAME")
	os.Unsetenv("FLEET_PASSWORD")
	os.Unsetenv("FLEET_API_KEY")
	os.Unsetenv("FLEET_BEARER_TOKEN")
	os.Unsetenv("FLEET_CA_CERTS")

	os.Setenv("FLEET_BEARER_TOKEN", "env-jwt-token")
	defer os.Unsetenv("FLEET_BEARER_TOKEN")

	kibanaCfg := kibanaOapiConfig{}
	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"fleet": providerSchema.GetFleetConnectionSchema(),
	}, map[string]interface{}{
		"fleet": []interface{}{
			map[string]interface{}{
				"endpoint":     "example.com/fleet",
				"bearer_token": "config-jwt-token",
			},
		},
	})

	fleetCfg, diags := newFleetConfigFromSDK(rd, kibanaCfg)

	require.Nil(t, diags)
	require.Equal(t, "env-jwt-token", fleetCfg.BearerToken)
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
				os.Setenv(key, val)
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

	os.Setenv("FLEET_BEARER_TOKEN", "env-jwt-token")
	defer os.Unsetenv("FLEET_BEARER_TOKEN")

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
