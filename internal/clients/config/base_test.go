package config

import (
	"net/http"
	"os"
	"testing"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestNewBaseConfigFromSDK(t *testing.T) {
	os.Unsetenv("ELASTICSEARCH_USERNAME")
	os.Unsetenv("ELASTICSEARCH_PASSWORD")
	os.Unsetenv("ELASTICSEARCH_API_KEY")

	rd := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"elasticsearch": providerSchema.GetEsConnectionSchema("elasticsearch", true),
	}, map[string]interface{}{
		"elasticsearch": []interface{}{
			map[string]interface{}{
				"username": "elastic",
				"password": "changeme",
			},
		},
	})

	baseCfg := newBaseConfigFromSDK(rd, "unit-testing", "elasticsearch")
	ua := "elasticstack-terraform-provider/unit-testing"
	require.Equal(t, baseConfig{
		Username:  "elastic",
		Password:  "changeme",
		UserAgent: ua,
		Header:    http.Header{"User-Agent": []string{ua}},
	}, baseCfg)
}

func TestNewBaseConfigFromFramework(t *testing.T) {
	os.Unsetenv("ELASTICSEARCH_USERNAME")
	os.Unsetenv("ELASTICSEARCH_PASSWORD")
	os.Unsetenv("ELASTICSEARCH_API_KEY")

	expectedUA := "elasticstack-terraform-provider/unit-testing"

	tests := []struct {
		name               string
		config             ProviderConfiguration
		expectedBaseConfig baseConfig
	}{
		{
			name: "with es config defined",
			config: ProviderConfiguration{
				Elasticsearch: []ElasticsearchConnection{
					{
						Username: types.StringValue("elastic"),
						Password: types.StringValue("changeme"),
						APIKey:   types.StringValue("apikey"),
					},
				},
			},
			expectedBaseConfig: baseConfig{
				Username:  "elastic",
				Password:  "changeme",
				ApiKey:    "apikey",
				UserAgent: expectedUA,
				Header:    http.Header{"User-Agent": []string{expectedUA}},
			},
		},
		{
			name: "should not set credentials if no configuration available",
			config: ProviderConfiguration{
				Elasticsearch: []ElasticsearchConnection{},
			},
			expectedBaseConfig: baseConfig{
				UserAgent: expectedUA,
				Header:    http.Header{"User-Agent": []string{expectedUA}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseCfg := newBaseConfigFromFramework(tt.config, "unit-testing")
			require.Equal(t, tt.expectedBaseConfig, baseCfg)
		})
	}
}
