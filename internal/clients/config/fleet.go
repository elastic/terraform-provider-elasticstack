package config

import (
	"context"
	"os"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type fleetConfig fleet.Config

func newFleetConfigFromSDK(d *schema.ResourceData, kibanaCfg kibanaOapiConfig) (fleetConfig, sdkdiags.Diagnostics) {
	config := kibanaCfg.toFleetConfig()

	// Set variables from resource config.
	if fleetDataRaw, ok := d.GetOk("fleet"); ok {
		fleetData, ok := fleetDataRaw.([]interface{})[0].(map[string]any)
		if !ok {
			diags := sdkdiags.Diagnostics{
				sdkdiags.Diagnostic{
					Severity: sdkdiags.Error,
					Summary:  "Unable to parse Fleet configuration",
					Detail:   "Fleet configuration data has not been configured correctly or is empty",
				},
			}
			return fleetConfig{}, diags
		}
		if v, ok := fleetData["endpoint"].(string); ok && v != "" {
			config.URL = v
		}
		if v, ok := fleetData["username"].(string); ok && v != "" {
			config.Username = v
		}
		if v, ok := fleetData["password"].(string); ok && v != "" {
			config.Password = v
		}
		if v, ok := fleetData["api_key"].(string); ok && v != "" {
			config.APIKey = v
		}
		if v, ok := fleetData["ca_certs"].([]interface{}); ok && len(v) > 0 {
			for _, elem := range v {
				if vStr, elemOk := elem.(string); elemOk {
					config.CACerts = append(config.CACerts, vStr)
				}
			}
		}
		if v, ok := fleetData["insecure"].(bool); ok {
			config.Insecure = v
		}
	}

	return config.withEnvironmentOverrides(), nil
}

func newFleetConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, kibanaCfg kibanaOapiConfig) (fleetConfig, fwdiags.Diagnostics) {
	config := kibanaCfg.toFleetConfig()

	if len(cfg.Fleet) > 0 {
		fleetCfg := cfg.Fleet[0]
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

		if !fleetCfg.Insecure.IsNull() && !fleetCfg.Insecure.IsUnknown() {
			config.Insecure = fleetCfg.Insecure.ValueBool()
		}

		var caCerts []string
		diags := fleetCfg.CACerts.ElementsAs(ctx, &caCerts, true)
		if diags.HasError() {
			return fleetConfig{}, diags
		}

		if len(caCerts) > 0 {
			config.CACerts = caCerts
		}
	}

	return config.withEnvironmentOverrides(), nil
}

func (c fleetConfig) withEnvironmentOverrides() fleetConfig {
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
	if v, ok := os.LookupEnv("FLEET_CA_CERTS"); ok {
		c.CACerts = strings.Split(v, ",")
	}

	return c
}
