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
	k.URL = withEnvironmentOverrideUnlessConfigured(k.URL, "KIBANA_ENDPOINT", PreferConfiguredKibanaEndpointEnvVar)
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
