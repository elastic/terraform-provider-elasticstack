package config

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiags "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type elasticsearchConfig elasticsearch.Config

func newElasticsearchConfigFromSDK(d *schema.ResourceData, base baseConfig, key string, useEnvAsDefault bool) (*elasticsearchConfig, sdkdiags.Diagnostics) {
	esConn, ok := d.GetOk(key)
	if !ok {
		return nil, nil
	}

	var diags sdkdiags.Diagnostics
	config := base.toElasticsearchConfig()

	// if defined, then we only have a single entry
	if es := esConn.([]interface{})[0]; es != nil {
		esConfig := es.(map[string]interface{})

		if endpoints, ok := esConfig["endpoints"]; ok && len(endpoints.([]interface{})) > 0 {
			var addrs []string
			for _, e := range endpoints.([]interface{}) {
				addrs = append(addrs, e.(string))
			}
			config.Addresses = addrs
		}

		if bearer_token, ok := esConfig["bearer_token"].(string); ok && bearer_token != "" {
			base.Header.Set("Authorization", bearer_token)
		}

		if es_client_authentication, ok := esConfig["es_client_authentication"].(string); ok && es_client_authentication != "" {
			base.Header.Set("ES-Client-Authentication", es_client_authentication)
		}

		if insecure, ok := esConfig["insecure"]; ok && insecure.(bool) {
			tlsClientConfig := config.ensureTLSClientConfig()
			tlsClientConfig.InsecureSkipVerify = true
		}

		if caFile, ok := esConfig["ca_file"]; ok && caFile.(string) != "" {
			caCert, err := os.ReadFile(caFile.(string))
			if err != nil {
				diags = append(diags, sdkdiags.Diagnostic{
					Severity: sdkdiags.Error,
					Summary:  "Unable to read CA File",
					Detail:   err.Error(),
				})
				return nil, diags
			}
			config.CACert = caCert
		}
		if caData, ok := esConfig["ca_data"]; ok && caData.(string) != "" {
			config.CACert = []byte(caData.(string))
		}

		if certFile, ok := esConfig["cert_file"]; ok && certFile.(string) != "" {
			if keyFile, ok := esConfig["key_file"]; ok && keyFile.(string) != "" {
				cert, err := tls.LoadX509KeyPair(certFile.(string), keyFile.(string))
				if err != nil {
					diags = append(diags, sdkdiags.Diagnostic{
						Severity: sdkdiags.Error,
						Summary:  "Unable to read certificate or key file",
						Detail:   err.Error(),
					})
					return nil, diags
				}
				tlsClientConfig := config.ensureTLSClientConfig()
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				diags = append(diags, sdkdiags.Diagnostic{
					Severity: sdkdiags.Error,
					Summary:  "Unable to read key file",
					Detail:   "Path to key file has not been configured or is empty",
				})
				return nil, diags
			}
		}
		if certData, ok := esConfig["cert_data"]; ok && certData.(string) != "" {
			if keyData, ok := esConfig["key_data"]; ok && keyData.(string) != "" {
				cert, err := tls.X509KeyPair([]byte(certData.(string)), []byte(keyData.(string)))
				if err != nil {
					diags = append(diags, sdkdiags.Diagnostic{
						Severity: sdkdiags.Error,
						Summary:  "Unable to parse certificate or key",
						Detail:   err.Error(),
					})
					return nil, diags
				}
				tlsClientConfig := config.ensureTLSClientConfig()
				tlsClientConfig.Certificates = []tls.Certificate{cert}
			} else {
				diags = append(diags, sdkdiags.Diagnostic{
					Severity: sdkdiags.Error,
					Summary:  "Unable to parse key",
					Detail:   "Key data has not been configured or is empty",
				})
				return nil, diags
			}
		}
	}

	if logging.IsDebugOrHigher() {
		config.EnableDebugLogger = true
		config.Logger = &debugLogger{Name: "elasticsearch"}
	}

	config = config.withEnvironmentOverrides()
	return &config, nil
}

func newElasticsearchConfigFromFramework(ctx context.Context, cfg ProviderConfiguration, base baseConfig) (*elasticsearchConfig, fwdiags.Diagnostics) {
	if len(cfg.Elasticsearch) == 0 {
		return nil, nil
	}

	config := base.toElasticsearchConfig()
	esConfig := cfg.Elasticsearch[0]

	var endpoints []string
	diags := esConfig.Endpoints.ElementsAs(ctx, &endpoints, true)
	if diags.HasError() {
		return nil, diags
	}

	if len(endpoints) > 0 {
		config.Addresses = endpoints
	}

	if esConfig.BearerToken.ValueString() != "" {
		config.Header.Set("Authorization", esConfig.BearerToken.ValueString())
		if esConfig.ESClientAuthentication.ValueString() != "" || esConfig.ESClientAuthentication.ValueString() != "null" {
			config.Header.Set("ES-Client-authentication", esConfig.ESClientAuthentication.ValueString())
		}
	}

	if esConfig.Insecure.ValueBool() {
		tlsClientConfig := config.ensureTLSClientConfig()
		tlsClientConfig.InsecureSkipVerify = true
	}

	if caFile := esConfig.CAFile.ValueString(); caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to read CA file", err.Error()))
			return nil, diags
		}
		config.CACert = caCert
	}
	if caData := esConfig.CAData.ValueString(); caData != "" {
		config.CACert = []byte(caData)
	}

	if certFile := esConfig.CertFile.ValueString(); certFile != "" {
		if keyFile := esConfig.KeyFile.ValueString(); keyFile != "" {
			cert, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				diags.Append(fwdiags.NewErrorDiagnostic("Unable to read certificate or key file", err.Error()))
				return nil, diags
			}
			tlsClientConfig := config.ensureTLSClientConfig()
			tlsClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to read key file", "Path to key file has not been configured or is empty"))
			return nil, diags
		}
	}
	if certData := esConfig.CertData.ValueString(); certData != "" {
		if keyData := esConfig.KeyData.ValueString(); keyData != "" {
			cert, err := tls.X509KeyPair([]byte(certData), []byte(keyData))
			if err != nil {
				diags.Append(fwdiags.NewErrorDiagnostic("Unable to parse certificate or key", err.Error()))
				return nil, diags
			}
			tlsClientConfig := config.ensureTLSClientConfig()
			tlsClientConfig.Certificates = []tls.Certificate{cert}
		} else {
			diags.Append(fwdiags.NewErrorDiagnostic("Unable to parse key", "Key data has not been configured or is empty"))
			return nil, diags
		}
	}

	if logging.IsDebugOrHigher() {
		config.EnableDebugLogger = true
		config.Logger = &debugLogger{Name: "elasticsearch"}
	}

	config = config.withEnvironmentOverrides()
	return &config, nil
}

func (c *elasticsearchConfig) ensureTLSClientConfig() *tls.Config {
	if c.Transport == nil {
		c.Transport = http.DefaultTransport.(*http.Transport)
	}
	if c.Transport.(*http.Transport).TLSClientConfig == nil {
		c.Transport.(*http.Transport).TLSClientConfig = &tls.Config{}
	}
	return c.Transport.(*http.Transport).TLSClientConfig
}

func (c elasticsearchConfig) withEnvironmentOverrides() elasticsearchConfig {
	if endpointsCSV, ok := os.LookupEnv("ELASTICSEARCH_ENDPOINTS"); ok {
		endpoints := make([]string, 0)
		for _, e := range strings.Split(endpointsCSV, ",") {
			endpoints = append(endpoints, strings.TrimSpace(e))
		}
		c.Addresses = endpoints
	}

	if insecure, ok := os.LookupEnv("ELASTICSEARCH_INSECURE"); ok {
		if insecureValue, err := strconv.ParseBool(insecure); err == nil {
			tlsClientConfig := c.ensureTLSClientConfig()
			tlsClientConfig.InsecureSkipVerify = insecureValue
		}
	}

	if bearerToken := os.Getenv("ELASTICSEARCH_BEARER_TOKEN"); bearerToken != "" {
		c.Header.Set("Authorization", bearerToken)
	}

	if esClientAuthentication := os.Getenv("ELASTICSEARCH_BEARER_TOKEN"); esClientAuthentication != "" {
		c.Header.Set("ES-Client-authentication", esClientAuthentication)
	}

	return c
}
