package config

import (
	"net/http"
	"os"
	"testing"

	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
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
	ua := buildUserAgent("unit-testing")
	require.Equal(t, baseConfig{
		Username:  "elastic",
		Password:  "changeme",
		UserAgent: ua,
		Header:    http.Header{"User-Agent": []string{ua}},
	}, baseCfg)
}
