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
)

type kibanaOapiConfig kibanaoapi.Config

func newKibanaOapiConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaOapiConfig, fwdiags.Diagnostics) {
	config, diags := buildKibanaOapiConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return kibanaOapiConfig{}, diags
	}

	config = config.withEnvironmentOverrides()

	if authMethodCount(kibanaoapi.Config(config)) > 1 {
		diags.AddWarning(
			"Multiple Kibana authentication methods configured",
			"More than one of username/password (username must be set), api_key, or bearer_token is set in "+
				"the resolved Kibana configuration. Only one will be used. Check your "+
				"provider configuration and environment variables for conflicting auth settings.",
		)
	}

	return config, diags
}

func newProviderKibanaOapiConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaOapiConfig, fwdiags.Diagnostics) {
	config, diags := buildKibanaOapiConfigFromFramework(ctx, cfg, base)
	if diags.HasError() {
		return kibanaOapiConfig{}, diags
	}

	// Apply the URL env override before fleet fallback so a fleet-derived URL does
	// not suppress KIBANA_ENDPOINT when TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT is set.
	config = config.withURLEnvironmentOverride()

	config, diags = config.withFleetBlockFallback(ctx, cfg)
	if diags.HasError() {
		return kibanaOapiConfig{}, diags
	}

	config = config.withNonURLEnvironmentOverrides()

	if authMethodCount(kibanaoapi.Config(config)) > 1 {
		diags.AddWarning(
			"Multiple Kibana authentication methods configured",
			"More than one of username/password (username must be set), api_key, or bearer_token is set in "+
				"the resolved Kibana configuration. Only one will be used. Check your "+
				"provider configuration and environment variables for conflicting auth settings.",
		)
	}

	return config, diags
}

func buildKibanaOapiConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaOapiConfig, fwdiags.Diagnostics) {
	config := base.toKibanaOapiConfig()

	if len(cfg.Kibana) > 0 {
		kibConfig := cfg.Kibana[0]

		applyAuthOverride(
			(*kibanaoapi.Config)(&config),
			kibConfig.Username.ValueString(),
			kibConfig.Password.ValueString(),
			kibConfig.APIKey.ValueString(),
			kibConfig.BearerToken.ValueString(),
		)
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

		if !kibConfig.Insecure.IsNull() && !kibConfig.Insecure.IsUnknown() {
			config.Insecure = kibConfig.Insecure.ValueBool()
		}
	}

	return config, nil
}

func (k kibanaOapiConfig) withFleetBlockFallback(ctx context.Context, cfg ProviderConfiguration) (kibanaOapiConfig, fwdiags.Diagnostics) {
	if len(cfg.Fleet) == 0 {
		return k, nil
	}

	fleetCfg := cfg.Fleet[0]

	kibanaHasAuth := k.Username != "" || k.Password != "" || k.APIKey != "" || k.BearerToken != ""
	if !kibanaHasAuth {
		if fleetCfg.Username.ValueString() != "" {
			k.Username = fleetCfg.Username.ValueString()
		}
		if fleetCfg.Password.ValueString() != "" {
			k.Password = fleetCfg.Password.ValueString()
		}
		if fleetCfg.APIKey.ValueString() != "" {
			k.APIKey = fleetCfg.APIKey.ValueString()
		}
		if fleetCfg.BearerToken.ValueString() != "" {
			k.BearerToken = fleetCfg.BearerToken.ValueString()
		}
	}
	if k.URL == "" && fleetCfg.Endpoint.ValueString() != "" {
		k.URL = fleetCfg.Endpoint.ValueString()
	}

	if len(k.CACerts) == 0 {
		var caCerts []string
		diags := fleetCfg.CACerts.ElementsAs(ctx, &caCerts, true)
		if diags.HasError() {
			return kibanaOapiConfig{}, diags
		}
		if len(caCerts) > 0 {
			k.CACerts = caCerts
		}
	}

	kibanaInsecureUnset := len(cfg.Kibana) == 0 || cfg.Kibana[0].Insecure.IsNull() || cfg.Kibana[0].Insecure.IsUnknown()
	if kibanaInsecureUnset && !fleetCfg.Insecure.IsNull() && !fleetCfg.Insecure.IsUnknown() {
		k.Insecure = fleetCfg.Insecure.ValueBool()
	}

	return k, nil
}

func (k kibanaOapiConfig) withURLEnvironmentOverride() kibanaOapiConfig {
	k.URL = withEnvironmentOverrideUnlessConfigured(k.URL, "KIBANA_ENDPOINT", PreferConfiguredKibanaEndpointEnvVar)
	return k
}

func (k kibanaOapiConfig) withEnvironmentOverrides() kibanaOapiConfig {
	k = k.withNonURLEnvironmentOverrides()
	k = k.withURLEnvironmentOverride()
	return k
}

func (k kibanaOapiConfig) withNonURLEnvironmentOverrides() kibanaOapiConfig {
	applyAuthEnvOverrides((*kibanaoapi.Config)(&k), "KIBANA")

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
