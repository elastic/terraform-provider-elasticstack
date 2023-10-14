package config

import (
	"context"
	"os"
	"strconv"

	"github.com/disaster37/go-kibana-rest/v8"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type kibanaConfig kibana.Config

func newKibanaConfigFromSDK(d *schema.ResourceData, base baseConfig) (kibanaConfig, sdkdiags.Diagnostics) {
	var diags sdkdiags.Diagnostics

	// Use ES details by default
	config := base.toKibanaConfig()
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

		if endpoints, ok := kibConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			// We're curently limited by the API to a single endpoint
			if endpoint := endpoints.([]interface{})[0]; endpoint != nil {
				config.Address = endpoint.(string)
			}
		}

		if insecure, ok := kibConfig["insecure"]; ok && insecure.(bool) {
			config.DisableVerifySSL = true
		}
	}

	return config.withEnvironmentOverrides(), nil
}

func newKibanaConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (kibanaConfig, fwdiags.Diagnostics) {
	config := base.toKibanaConfig()

	if len(cfg.Kibana) > 0 {
		kibConfig := cfg.Kibana[0]
		if kibConfig.Username.ValueString() != "" {
			config.Username = kibConfig.Username.ValueString()
		}
		if kibConfig.Password.ValueString() != "" {
			config.Password = kibConfig.Password.ValueString()
		}
		var endpoints []string
		diags := kibConfig.Endpoints.ElementsAs(ctx, &endpoints, true)
		if diags.HasError() {
			return kibanaConfig{}, diags
		}

		if len(endpoints) > 0 {
			config.Address = endpoints[0]
		}

		config.DisableVerifySSL = kibConfig.Insecure.ValueBool()
	}

	return config.withEnvironmentOverrides(), nil
}

func (k kibanaConfig) withEnvironmentOverrides() kibanaConfig {
	k.Username = withEnvironmentOverride(k.Username, "KIBANA_USERNAME")
	k.Password = withEnvironmentOverride(k.Password, "KIBANA_PASSWORD")
	k.Address = withEnvironmentOverride(k.Address, "KIBANA_ENDPOINT")

	if insecure, ok := os.LookupEnv("KIBANA_INSECURE"); ok {
		if insecureValue, err := strconv.ParseBool(insecure); err == nil {
			k.DisableVerifySSL = insecureValue
		}
	}

	return k
}

func (k kibanaConfig) toFleetConfig() fleetConfig {
	return fleetConfig{
		URL:      k.Address,
		Username: k.Username,
		Password: k.Password,
		Insecure: k.DisableVerifySSL,
	}
}
