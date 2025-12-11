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
		{
			name: "should fallback to kibana username/password when no fleet credentials provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								// No username/password provided in fleet config
								"ca_certs": []interface{}{"internal"},
								"insecure": false,
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "kibana-user",     // Falls back to kibana credentials
						Password: "kibana-password", // Falls back to kibana credentials
						CACerts:  []string{"internal"},
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should fallback to kibana api_key when no fleet credentials provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: true,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								// No api_key provided in fleet config
								"insecure": true,
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "kibana-api-key-123", // Falls back to kibana api_key
						Insecure: true,
					},
				}
			},
		},
		{
			name: "should not override fleet credentials when they are explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"username": "fleet-user",
								"password": "fleet-password",
								"insecure": false,
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "fleet-user",     // Uses fleet-specific credentials
						Password: "fleet-password", // Uses fleet-specific credentials
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not override fleet api_key when explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"api_key":  "fleet-api-key-456",
								"insecure": false,
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "fleet-api-key-456", // Uses fleet-specific api_key
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not add kibana api key when fleet credentials are explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"username": "fleet-user",
								"password": "fleet-password",
								"insecure": false,
							},
						},
					},
					env: map[string]string{},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "fleet-user",     // Uses fleet-specific credentials
						Password: "fleet-password", // Uses fleet-specific credentials
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not add kibana credentials when fleet api key is explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					resourceData: map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"endpoint": "example.com/fleet",
								"insecure": false,
								"api_key":  "fleet-api-key-123",
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "fleet-api-key-123",
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
		{
			name: "should fallback to kibana username/password when no fleet credentials provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("example.com/fleet"),
								// No username/password provided in fleet config
								CACerts: types.ListValueMust(types.StringType, []attr.Value{
									types.StringValue("internal"),
								}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "kibana-user",     // Falls back to kibana credentials
						Password: "kibana-password", // Falls back to kibana credentials
						CACerts:  []string{"internal"},
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should fallback to kibana api_key when no fleet credentials provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: true,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Endpoint: types.StringValue("example.com/fleet"),
								// No api_key provided in fleet config
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(true),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "kibana-api-key-123", // Falls back to kibana api_key
						Insecure: true,
					},
				}
			},
		},
		{
			name: "should not override fleet credentials when they are explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								Password: types.StringValue("fleet-password"),
								Endpoint: types.StringValue("example.com/fleet"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "fleet-user",     // Uses fleet-specific credentials
						Password: "fleet-password", // Uses fleet-specific credentials
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not override fleet api_key when explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								APIKey:   types.StringValue("fleet-api-key-456"),
								Endpoint: types.StringValue("example.com/fleet"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "fleet-api-key-456", // Uses fleet-specific api_key
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not add kibana api key when fleet credentials are explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					APIKey:   "kibana-api-key-123",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								Username: types.StringValue("fleet-user"),
								Password: types.StringValue("fleet-password"),
								Endpoint: types.StringValue("example.com/fleet"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					env: map[string]string{},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						Username: "fleet-user",     // Uses fleet-specific credentials
						Password: "fleet-password", // Uses fleet-specific credentials
						Insecure: false,
					},
				}
			},
		},
		{
			name: "should not add kibana credentials when fleet api key is explicitly provided",
			args: func() args {
				kibanaCfg := kibanaOapiConfig{
					URL:      "example.com/kibana",
					Username: "kibana-user",
					Password: "kibana-password",
					Insecure: false,
				}

				return args{
					kibanaCfg: kibanaCfg,
					providerConfig: ProviderConfiguration{
						Fleet: []FleetConnection{
							{
								APIKey:   types.StringValue("fleet-api-key-456"),
								Endpoint: types.StringValue("example.com/fleet"),
								CACerts:  types.ListValueMust(types.StringType, []attr.Value{}),
								Insecure: types.BoolValue(false),
							},
						},
					},
					expectedConfig: fleetConfig{
						URL:      "example.com/fleet",
						APIKey:   "fleet-api-key-456",
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
