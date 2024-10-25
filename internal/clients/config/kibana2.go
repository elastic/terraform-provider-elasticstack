package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibana2"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type kibana2Config kibana2.Config

func newKibana2ConfigFromSDK(d *schema.ResourceData, base baseConfig) (kibana2Config, sdkdiags.Diagnostics) {
	var diags sdkdiags.Diagnostics

	// Use ES details by default
	config := base.toKibana2Config()
	kibConn, ok := d.GetOk("kibana")
	if !ok {
		return config, diags
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
			config.APIKey = apiKey.(string)
		}

		if endpoints, ok := kibConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			// We're curently limited by the API to a single endpoint
			if endpoint := endpoints.([]interface{})[0]; endpoint != nil {
				config.URL = endpoint.(string)
			}
		}

		if caCerts, ok := kibConfig["ca_certs"].([]interface{}); ok && len(caCerts) > 0 {
			for _, elem := range caCerts {
				if vStr, elemOk := elem.(string); elemOk {
					config.CACerts = append(config.CACerts, vStr)
				}
			}
		}

		if insecure, ok := kibConfig["insecure"]; ok && insecure.(bool) {
			config.Insecure = true
		}
	}

	return config.withEnvironmentOverrides(), nil
}

func newKibana2ConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibana2Config, fwdiags.Diagnostics) {
	config := base.toKibana2Config()

	if len(cfg.Kibana) > 0 {
		kibConfig := cfg.Kibana[0]
		if kibConfig.Username.ValueString() != "" {
			config.Username = kibConfig.Username.ValueString()
		}
		if kibConfig.Password.ValueString() != "" {
			config.Password = kibConfig.Password.ValueString()
		}
		if kibConfig.ApiKey.ValueString() != "" {
			config.APIKey = kibConfig.ApiKey.ValueString()
		}
		var endpoints []string
		diags := kibConfig.Endpoints.ElementsAs(ctx, &endpoints, true)

		var cas []string
		diags.Append(kibConfig.CACerts.ElementsAs(ctx, &cas, true)...)
		if diags.HasError() {
			return kibana2Config{}, diags
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

func (k kibana2Config) withEnvironmentOverrides() kibana2Config {
	k.Username = withEnvironmentOverride(k.Username, "KIBANA_USERNAME")
	k.Password = withEnvironmentOverride(k.Password, "KIBANA_PASSWORD")
	k.APIKey = withEnvironmentOverride(k.APIKey, "KIBANA_API_KEY")
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

func (k kibana2Config) toFleetConfig() fleetConfig {
	return fleetConfig(k)
}
