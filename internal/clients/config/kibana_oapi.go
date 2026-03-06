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
	"strconv"
	"strings"

	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type kibanaOapiConfig kibanaoapi.Config

func newKibanaOapiConfigFromSDK(d *schema.ResourceData, base baseConfig) (kibanaOapiConfig, sdkdiags.Diagnostics) {
	// Use ES details by default
	config := base.toKibanaOapiConfig()
	kibConn, ok := d.GetOk("kibana")
	if !ok {
		return config, nil
	}

	kibConnList, ok := kibConn.([]any)
	if !ok || len(kibConnList) == 0 {
		return config, sdkdiags.Errorf("invalid provider configuration: kibana must be a non-empty list")
	}

	// if defined, then we only have a single entry
	if kib := kibConnList[0]; kib != nil {
		kibConfig, ok := kib.(map[string]any)
		if !ok {
			return config, sdkdiags.Errorf("invalid provider configuration: kibana[0] must be an object")
		}

		if usernameRaw, usernameOk := kibConfig["username"]; usernameOk {
			switch v := usernameRaw.(type) {
			case string:
				if v != "" {
					config.Username = v
				}
			case nil:
			default:
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.username must be a string")
			}
		}

		if passwordRaw, passwordOk := kibConfig["password"]; passwordOk {
			switch v := passwordRaw.(type) {
			case string:
				if v != "" {
					config.Password = v
				}
			case nil:
			default:
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.password must be a string")
			}
		}

		if apiKeyRaw, apiKeyOk := kibConfig["api_key"]; apiKeyOk {
			switch v := apiKeyRaw.(type) {
			case string:
				if v != "" {
					config.APIKey = v
				}
			case nil:
			default:
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.api_key must be a string")
			}
		}

		if bearerTokenRaw, bearerTokenOk := kibConfig["bearer_token"]; bearerTokenOk {
			switch v := bearerTokenRaw.(type) {
			case string:
				if v != "" {
					config.BearerToken = v
				}
			case nil:
			default:
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.bearer_token must be a string")
			}
		}

		if endpointsRaw, endpointsOk := kibConfig["endpoints"]; endpointsOk {
			endpointsList, ok := endpointsRaw.([]any)
			if !ok {
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.endpoints must be a list")
			}
			if len(endpointsList) > 0 {
				// We're curently limited by the API to a single endpoint
				if endpoint := endpointsList[0]; endpoint != nil {
					endpointStr, ok := endpoint.(string)
					if !ok {
						return config, sdkdiags.Errorf("invalid provider configuration: kibana.endpoints must be a list of strings")
					}
					config.URL = endpointStr
				}
			}
		}

		if caCertsRaw, caCertsOk := kibConfig["ca_certs"]; caCertsOk {
			caCerts, ok := caCertsRaw.([]any)
			if !ok {
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.ca_certs must be a list")
			}
			for _, elem := range caCerts {
				if elem == nil {
					continue
				}
				vStr, ok := elem.(string)
				if !ok {
					return config, sdkdiags.Errorf("invalid provider configuration: kibana.ca_certs must be a list of strings")
				}
				config.CACerts = append(config.CACerts, vStr)
			}
		}

		if insecureRaw, insecureOk := kibConfig["insecure"]; insecureOk {
			insecure, ok := insecureRaw.(bool)
			if !ok {
				return config, sdkdiags.Errorf("invalid provider configuration: kibana.insecure must be a bool")
			}
			if insecure {
				config.Insecure = true
			}
		}
	}

	return config.withEnvironmentOverrides(), nil
}

func newKibanaOapiConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaOapiConfig, fwdiags.Diagnostics) {
	config := base.toKibanaOapiConfig()

	if len(cfg.Kibana) > 0 {
		kibConfig := cfg.Kibana[0]
		if kibConfig.Username.ValueString() != "" {
			config.Username = kibConfig.Username.ValueString()
		}
		if kibConfig.Password.ValueString() != "" {
			config.Password = kibConfig.Password.ValueString()
		}
		if kibConfig.APIKey.ValueString() != "" {
			config.APIKey = kibConfig.APIKey.ValueString()
		}
		if kibConfig.BearerToken.ValueString() != "" {
			config.BearerToken = kibConfig.BearerToken.ValueString()
		}
		var endpoints []string
		diags := kibConfig.Endpoints.ElementsAs(ctx, &endpoints, true)

		var cas []string
		diags.Append(kibConfig.CACerts.ElementsAs(ctx, &cas, true)...)
		if diags.HasError() {
			return kibanaOapiConfig{}, diags
		}

		if len(endpoints) > 0 {
			config.URL = endpoints[0]
		}

		if len(cas) > 0 {
			config.CACerts = cas
		}

		config.Insecure = kibConfig.Insecure.ValueBool()
	}

	return config.withEnvironmentOverrides(), nil
}

func (k kibanaOapiConfig) withEnvironmentOverrides() kibanaOapiConfig {
	k.Username = withEnvironmentOverride(k.Username, "KIBANA_USERNAME")
	k.Password = withEnvironmentOverride(k.Password, "KIBANA_PASSWORD")
	k.APIKey = withEnvironmentOverride(k.APIKey, "KIBANA_API_KEY")
	k.BearerToken = withEnvironmentOverride(k.BearerToken, "KIBANA_BEARER_TOKEN")
	k.URL = withEnvironmentOverride(k.URL, "KIBANA_ENDPOINT")
	if caCerts, ok := os.LookupEnv("KIBANA_CA_CERTS"); ok {
		k.CACerts = strings.Split(caCerts, ",")
	}

	if insecure, ok := os.LookupEnv("KIBANA_INSECURE"); ok {
		if insecureValue, err := strconv.ParseBool(insecure); err == nil {
			k.Insecure = insecureValue
		}
	}

	return k
}

func (k kibanaOapiConfig) toFleetConfig() fleetConfig {
	return fleetConfig(k)
}
