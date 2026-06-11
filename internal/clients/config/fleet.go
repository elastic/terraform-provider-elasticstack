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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

type fleetConfig fleet.Config

func newFleetConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, kibanaCfg kibanaOapiConfig) (fleetConfig, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics
	config := kibanaCfg.toFleetConfig()

	if len(cfg.Fleet) > 0 {
		fleetCfg := cfg.Fleet[0]

		fleetUsesBasicAuth := fleetCfg.Username.ValueString() != "" || fleetCfg.Password.ValueString() != ""
		fleetUsesAPIKey := fleetCfg.APIKey.ValueString() != ""
		fleetUsesBearer := fleetCfg.BearerToken.ValueString() != ""

		switch {
		case fleetUsesBearer:
			clearConflictingAuth((*kibanaoapi.Config)(&config), authMethodBearerToken)
		case fleetUsesAPIKey:
			clearConflictingAuth((*kibanaoapi.Config)(&config), authMethodAPIKey)
		case fleetUsesBasicAuth:
			clearConflictingAuth((*kibanaoapi.Config)(&config), authMethodBasicAuth)
		}

		if fleetCfg.Username.ValueString() != "" {
			config.Username = fleetCfg.Username.ValueString()
		}
		if fleetCfg.Password.ValueString() != "" {
			config.Password = fleetCfg.Password.ValueString()
		}
		if fleetCfg.Endpoint.ValueString() != "" {
			config.URL = fleetCfg.Endpoint.ValueString()
		}
		if fleetCfg.APIKey.ValueString() != "" {
			config.APIKey = fleetCfg.APIKey.ValueString()
		}
		if fleetCfg.BearerToken.ValueString() != "" {
			config.BearerToken = fleetCfg.BearerToken.ValueString()
		}

		if !fleetCfg.Insecure.IsNull() && !fleetCfg.Insecure.IsUnknown() {
			config.Insecure = fleetCfg.Insecure.ValueBool()
		}

		var caCerts []string
		diags.Append(fleetCfg.CACerts.ElementsAs(ctx, &caCerts, true)...)
		if diags.HasError() {
			return fleetConfig{}, diags
		}

		if len(caCerts) > 0 {
			config.CACerts = caCerts
		}
	}

	config = config.withEnvironmentOverrides()

	if authMethodCount(kibanaoapi.Config(config)) > 1 {
		diags.AddWarning(
			"Multiple Fleet authentication methods configured",
			"More than one of username/password, api_key, or bearer_token is set in "+
				"the resolved Fleet configuration. Only one will be used. Check your "+
				"Fleet environment variables for conflicting auth settings.",
		)
	}

	return config, diags
}

func (c fleetConfig) withEnvironmentOverrides() fleetConfig {
	_, hasUser := os.LookupEnv("FLEET_USERNAME")
	_, hasPass := os.LookupEnv("FLEET_PASSWORD")
	_, hasKey := os.LookupEnv("FLEET_API_KEY")
	_, hasBearer := os.LookupEnv("FLEET_BEARER_TOKEN")

	switch {
	case hasBearer:
		clearConflictingAuth((*kibanaoapi.Config)(&c), authMethodBearerToken)
	case hasKey:
		clearConflictingAuth((*kibanaoapi.Config)(&c), authMethodAPIKey)
	case hasUser || hasPass:
		clearConflictingAuth((*kibanaoapi.Config)(&c), authMethodBasicAuth)
	}

	if v, ok := os.LookupEnv("FLEET_ENDPOINT"); ok {
		c.URL = v
	}
	if v, ok := os.LookupEnv("FLEET_USERNAME"); ok {
		c.Username = v
	}
	if v, ok := os.LookupEnv("FLEET_PASSWORD"); ok {
		c.Password = v
	}
	if v, ok := os.LookupEnv("FLEET_API_KEY"); ok {
		c.APIKey = v
	}
	if v, ok := os.LookupEnv("FLEET_BEARER_TOKEN"); ok {
		c.BearerToken = v
	}
	if v, ok := os.LookupEnv("FLEET_CA_CERTS"); ok {
		c.CACerts = strings.Split(v, ",")
	}

	return c
}
