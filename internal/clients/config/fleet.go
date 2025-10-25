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

// Structure to keep track of which keys were explicitly set in the config.
// This allows us to determine the difference between explicitly set empty
// values and values that were not set at all. Building this intermediate
// representation allows for compatibility with plugin framework and sdkv2.
type fleetConfigKeys struct {
	URL      bool
	Username bool
	Password bool
	APIKey   bool
	Insecure bool
	CACerts  bool
}

func newFleetConfigFromSDK(d *schema.ResourceData, kibanaCfg kibanaOapiConfig) (fleetConfig, sdkdiags.Diagnostics) {
	config := fleetConfig{}

	// Keep track of keys that are explicitly set in the config.
	knownKeys := fleetConfigKeys{}

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
			knownKeys.URL = true
		}
		if v, ok := fleetData["username"].(string); ok && v != "" {
			config.Username = v
			knownKeys.Username = true
		}
		if v, ok := fleetData["password"].(string); ok && v != "" {
			config.Password = v
			knownKeys.Password = true
		}
		if v, ok := fleetData["api_key"].(string); ok && v != "" {
			config.APIKey = v
			knownKeys.APIKey = true
		}
		if v, ok := fleetData["ca_certs"].([]interface{}); ok && len(v) > 0 {
			for _, elem := range v {
				if vStr, elemOk := elem.(string); elemOk {
					config.CACerts = append(config.CACerts, vStr)
				}
			}
			knownKeys.CACerts = true
		}
		if v, ok := fleetData["insecure"].(bool); ok {
			config.Insecure = v
			knownKeys.Insecure = true
		}
	}

	return config.withEnvironmentOverrides().withDefaultsApplied(kibanaCfg.toFleetConfig(), knownKeys), nil
}

func newFleetConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, kibanaCfg kibanaOapiConfig) (fleetConfig, fwdiags.Diagnostics) {
	config := fleetConfig{}
	// Keep track of keys that are explicitly set in the config
	knownKeys := fleetConfigKeys{}

	if len(cfg.Fleet) > 0 {
		fleetCfg := cfg.Fleet[0]
		if fleetCfg.Username.ValueString() != "" {
			config.Username = fleetCfg.Username.ValueString()
			knownKeys.Username = true
		}
		if fleetCfg.Password.ValueString() != "" {
			config.Password = fleetCfg.Password.ValueString()
			knownKeys.Password = true
		}
		if fleetCfg.Endpoint.ValueString() != "" {
			config.URL = fleetCfg.Endpoint.ValueString()
			knownKeys.URL = true
		}
		if fleetCfg.APIKey.ValueString() != "" {
			config.APIKey = fleetCfg.APIKey.ValueString()
			knownKeys.APIKey = true
		}

		if !fleetCfg.Insecure.IsNull() && !fleetCfg.Insecure.IsUnknown() {
			config.Insecure = fleetCfg.Insecure.ValueBool()
			knownKeys.Insecure = true
		}

		var caCerts []string
		diags := fleetCfg.CACerts.ElementsAs(ctx, &caCerts, true)
		if diags.HasError() {
			return fleetConfig{}, diags
		}

		if len(caCerts) > 0 {
			config.CACerts = caCerts
			knownKeys.CACerts = true
		}
	}

	return config.withEnvironmentOverrides().withDefaultsApplied(kibanaCfg.toFleetConfig(), knownKeys), nil
}

func (config fleetConfig) withDefaultsApplied(defaults fleetConfig, knownKeys fleetConfigKeys) fleetConfig {

	// Apply defaults for non-auth fields
	if !knownKeys.URL {
		config.URL = defaults.URL
	}
	if !knownKeys.Insecure {
		config.Insecure = defaults.Insecure
	}

	if !knownKeys.CACerts {
		config.CACerts = defaults.CACerts
	}

	// Handle auth defaults. APIKey and Username are mutually exclusive. If one is already set don't apply any auth defaults
	if knownKeys.APIKey || knownKeys.Username {
		return config
	}

	// Only apply a single provided auth default
	if defaults.APIKey != "" {
		config.APIKey = defaults.APIKey
	} else if defaults.Username != "" {
		config.Username = defaults.Username
		config.Password = defaults.Password
	}

	return config
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
