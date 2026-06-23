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

		applyAuthOverride(
			(*kibanaoapi.Config)(&config),
			fleetCfg.Username.ValueString(),
			fleetCfg.Password.ValueString(),
			fleetCfg.APIKey.ValueString(),
			fleetCfg.BearerToken.ValueString(),
		)

		if fleetCfg.Endpoint.ValueString() != "" {
			config.URL = fleetCfg.Endpoint.ValueString()
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
		addMultipleAuthWarning(&diags, "Fleet", "Fleet environment variables")
	}

	return config, diags
}

func (c fleetConfig) withEnvironmentOverrides() fleetConfig {
	applyAuthEnvOverrides((*kibanaoapi.Config)(&c), "FLEET")

	if v, ok := os.LookupEnv("FLEET_ENDPOINT"); ok {
		c.URL = v
	}
	if v, ok := os.LookupEnv("FLEET_CA_CERTS"); ok {
		c.CACerts = strings.Split(v, ",")
	}

	return c
}
