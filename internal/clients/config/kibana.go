package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/disaster37/go-kibana-rest/v8"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type kibanaConfig kibana.Config

func newKibanaConfigFromSDK(d *schema.ResourceData, base baseConfig) (kibanaConfig, sdkdiags.Diagnostics) {
	var diags sdkdiags.Diagnostics

	// Use ES details by default
	config := kibanaConfig{}

	kibConn, ok := d.GetOk("kibana")
	if !ok {
		return config.withDefaultsApplied(base.toKibanaConfig()), diags
	}

	// if defined, then we only have a single entry
	if kib := kibConn.([]interface{})[0]; kib != nil {
		kibConfig := kib.(map[string]interface{})

		if username, ok := kibConfig["username"]; ok && username != "" {
			config.Username = username.(string)
		}
		if password, ok := kibConfig["password"]; ok && password != "" {
			config.Password = password.(string)
		}

		if apiKey, ok := kibConfig["api_key"]; ok && apiKey != "" {
			config.ApiKey = apiKey.(string)
		}

		if endpoints, ok := kibConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			// We're curently limited by the API to a single endpoint
			if endpoint := endpoints.([]interface{})[0]; endpoint != nil {
				config.Address = endpoint.(string)
			}
		}

		if caCerts, ok := kibConfig["ca_certs"].([]interface{}); ok && len(caCerts) > 0 {
			for _, elem := range caCerts {
				if vStr, elemOk := elem.(string); elemOk {
					config.CAs = append(config.CAs, vStr)
				}
			}
		}

		if insecure, ok := kibConfig["insecure"]; ok && insecure.(bool) {
			config.DisableVerifySSL = true
		}
	}

	return config.withEnvironmentOverrides().withDefaultsApplied(base.toKibanaConfig()), nil
}

func newKibanaConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaConfig, fwdiags.Diagnostics) {
	config := kibanaConfig{}

	if len(cfg.Kibana) > 0 {
		kibConfig := cfg.Kibana[0]
		if kibConfig.Username.ValueString() != "" {
			config.Username = kibConfig.Username.ValueString()
		}
		if kibConfig.Password.ValueString() != "" {
			config.Password = kibConfig.Password.ValueString()
		}
		if kibConfig.ApiKey.ValueString() != "" {
			config.ApiKey = kibConfig.ApiKey.ValueString()
		}
		var endpoints []string
		diags := kibConfig.Endpoints.ElementsAs(ctx, &endpoints, true)

		var cas []string
		diags.Append(kibConfig.CACerts.ElementsAs(ctx, &cas, true)...)
		if diags.HasError() {
			return kibanaConfig{}, diags
		}

		if len(endpoints) > 0 {
			config.Address = endpoints[0]
		}

		if len(cas) > 0 {
			config.CAs = cas
		}

		config.DisableVerifySSL = kibConfig.Insecure.ValueBool()
	}
	return config.withEnvironmentOverrides().withDefaultsApplied(base.toKibanaConfig()), nil
}

func (config kibanaConfig) withDefaultsApplied(defaults kibanaConfig) kibanaConfig {

	// Apply defaults for non-auth fields
	if config.Address == "" {
		config.Address = defaults.Address
	}
	if config.DisableVerifySSL == false {
		config.DisableVerifySSL = defaults.DisableVerifySSL
	}
	if config.CAs == nil {
		config.CAs = defaults.CAs
	}

	// Handle auth defaults. ApiKey and Username are mutually exclusive. If one is already set don't apply any auth defaults
	if config.ApiKey != "" || config.Username != "" {
		return config
	}

	if defaults.ApiKey != "" {
		config.ApiKey = defaults.ApiKey
	} else if defaults.Username != "" {
		config.Username = defaults.Username
		config.Password = defaults.Password
	}

	return config

}

func (k kibanaConfig) withEnvironmentOverrides() kibanaConfig {
	k.Username = withEnvironmentOverride(k.Username, "KIBANA_USERNAME")
	k.Password = withEnvironmentOverride(k.Password, "KIBANA_PASSWORD")
	k.ApiKey = withEnvironmentOverride(k.ApiKey, "KIBANA_API_KEY")
	k.Address = withEnvironmentOverride(k.Address, "KIBANA_ENDPOINT")
	if caCerts, ok := os.LookupEnv("KIBANA_CA_CERTS"); ok {
		k.CAs = strings.Split(caCerts, ",")
	}

	if insecure, ok := os.LookupEnv("KIBANA_INSECURE"); ok {
		if insecureValue, err := strconv.ParseBool(insecure); err == nil {
			k.DisableVerifySSL = insecureValue
		}
	}

	return k
}
