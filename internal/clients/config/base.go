package config

import (
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type baseConfig struct {
	Username  string
	Password  string
	ApiKey    string
	UserAgent string
	Header    http.Header
}

func newBaseConfigFromSDK(d *schema.ResourceData, version string, esKey string) baseConfig {
	userAgent := buildUserAgent(version)
	baseConfig := baseConfig{
		UserAgent: userAgent,
		Header:    http.Header{"User-Agent": []string{userAgent}},
	}

	if elasticDataRaw, ok := d.GetOk("elasticsearch"); ok {
		elasticDataRaw, ok := elasticDataRaw.([]interface{})[0].(map[string]any)
		if token_value, ok := elasticDataRaw["bearer_token"].(string); ok && token_value != "" {
			token_value = "Bearer " + token_value
			if es_client_authentication, ok := elasticDataRaw["es_client_authentication"].(string); ok && es_client_authentication != "" {
				baseConfig.Header = http.Header{"Authorization": []string{token_value}, "ES-Client-Authentication": []string{es_client_authentication}, "User-Agent": []string{userAgent}}
			} else {
				baseConfig.Header = http.Header{"Authorization": []string{token_value}, "User-Agent": []string{userAgent}}
			}
		}
	} else {
		baseConfig.Header = http.Header{"User-Agent": []string{userAgent}}
	}

	if esConn, ok := d.GetOk(esKey); ok {
		if resource := esConn.([]interface{})[0]; resource != nil {
			config := resource.(map[string]interface{})

			if apiKey, ok := config["api_key"]; ok && apiKey != "" {
				baseConfig.ApiKey = apiKey.(string)
			} else {
				if username, ok := config["username"]; ok {
					baseConfig.Username = username.(string)
				}
				if password, ok := config["password"]; ok {
					baseConfig.Password = password.(string)
				}
			}
		}
	}

	return baseConfig.withEnvironmentOverrides()
}

func newBaseConfigFromFramework(config ProviderConfiguration, version string) baseConfig {
	userAgent := buildUserAgent(version)
	baseConfig := baseConfig{
		UserAgent: userAgent,
		Header:    http.Header{"User-Agent": []string{userAgent}},
	}

	if len(config.Elasticsearch) > 0 {
		esConfig := config.Elasticsearch[0]
		baseConfig.Username = esConfig.Username.ValueString()
		baseConfig.Password = esConfig.Password.ValueString()
		baseConfig.ApiKey = esConfig.APIKey.ValueString()
		if esConfig.BearerToken.ValueString() != "" || esConfig.BearerToken.ValueString() != "null" {
			if esConfig.EsClientAuthentication.ValueString() != "" || esConfig.EsClientAuthentication.ValueString() != "null" {
				baseConfig.Header = http.Header{"Authorization": []string{esConfig.BearerToken.ValueString()}, "ES-Client-Authentication": []string{esConfig.EsClientAuthentication.ValueString()}, "User-Agent": []string{userAgent}}
			} else {
			baseConfig.Header = http.Header{"Authorization": []string{esConfig.BearerToken.ValueString()}, "User-Agent": []string{userAgent}}
			}
		}
	}

	return baseConfig.withEnvironmentOverrides()
}

func (b baseConfig) withEnvironmentOverrides() baseConfig {
	b.Username = withEnvironmentOverride(b.Username, "ELASTICSEARCH_USERNAME")
	b.Password = withEnvironmentOverride(b.Password, "ELASTICSEARCH_PASSWORD")
	b.ApiKey = withEnvironmentOverride(b.ApiKey, "ELASTICSEARCH_API_KEY")

	return b
}

func (b baseConfig) toKibanaConfig() kibanaConfig {
	return kibanaConfig{
		Username: b.Username,
		Password: b.Password,
	}
}

func (b baseConfig) toElasticsearchConfig() elasticsearchConfig {
	return elasticsearchConfig{
		Header:   b.Header,
		Username: b.Username,
		Password: b.Password,
		APIKey:   b.ApiKey,
	}
}

func withEnvironmentOverride(currentValue, envOverrideKey string) string {
	if envValue, ok := os.LookupEnv(envOverrideKey); ok {
		return envValue
	}

	return currentValue
}

func buildUserAgent(version string) string {
	return fmt.Sprintf("elasticstack-terraform-provider/%s", version)
}
