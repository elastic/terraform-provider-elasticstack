package config

import (
	"fmt"
	"net/http"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type baseConfig struct {
	Username    string
	Password    string
	APIKey      string
	BearerToken string
	UserAgent   string
	Header      http.Header
}

func newBaseConfigFromSDK(d *schema.ResourceData, version string, esKey string) baseConfig {
	userAgent := buildUserAgent(version)
	baseConfig := baseConfig{
		UserAgent: userAgent,
		Header:    http.Header{"User-Agent": []string{userAgent}},
	}

	if esConn, ok := d.GetOk(esKey); ok {
		if resource := esConn.([]any)[0]; resource != nil {
			config := resource.(map[string]any)

			if bearerToken, ok := config["bearer_token"]; ok && bearerToken != "" {
				baseConfig.BearerToken = bearerToken.(string)
			} else if apiKey, ok := config["api_key"]; ok && apiKey != "" {
				baseConfig.APIKey = apiKey.(string)
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
		baseConfig.APIKey = esConfig.APIKey.ValueString()
		baseConfig.BearerToken = esConfig.BearerToken.ValueString()
	}

	return baseConfig.withEnvironmentOverrides()
}

func (b baseConfig) withEnvironmentOverrides() baseConfig {
	b.Username = withEnvironmentOverride(b.Username, "ELASTICSEARCH_USERNAME")
	b.Password = withEnvironmentOverride(b.Password, "ELASTICSEARCH_PASSWORD")
	b.APIKey = withEnvironmentOverride(b.APIKey, "ELASTICSEARCH_API_KEY")
	b.BearerToken = withEnvironmentOverride(b.BearerToken, "ELASTICSEARCH_BEARER_TOKEN")

	return b
}

func (b baseConfig) toKibanaConfig() kibanaConfig {
	return kibanaConfig{
		Username:    b.Username,
		Password:    b.Password,
		ApiKey:      b.APIKey,
		BearerToken: b.BearerToken,
	}
}

func (b baseConfig) toKibanaOapiConfig() kibanaOapiConfig {
	return kibanaOapiConfig{
		Username:    b.Username,
		Password:    b.Password,
		APIKey:      b.APIKey,
		BearerToken: b.BearerToken,
	}
}

func (b baseConfig) toElasticsearchConfig() elasticsearchConfig {
	return elasticsearchConfig{
		config: elasticsearch.Config{
			Header:   b.Header.Clone(),
			Username: b.Username,
			Password: b.Password,
			APIKey:   b.APIKey,
		},
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
