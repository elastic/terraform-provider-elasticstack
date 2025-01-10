package acctest

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var Providers map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	providerServerFactory, err := provider.ProtoV6ProviderServerFactory(context.Background(), "dev")
	if err != nil {
		log.Fatal(err)
	}
	Providers = map[string]func() (tfprotov6.ProviderServer, error){
		"elasticstack": func() (tfprotov6.ProviderServer, error) {
			return providerServerFactory(), nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, elasticsearchEndpointsOk := os.LookupEnv("ELASTICSEARCH_ENDPOINTS")
	_, kibanaEndpointOk := os.LookupEnv("KIBANA_ENDPOINT")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")
	_, apiKeyOk := os.LookupEnv("ELASTICSEARCH_API_KEY")
	_, kbUserOk := os.LookupEnv("KIBANA_USERNAME")
	_, kbPassOk := os.LookupEnv("KIBANA_PASSWORD")
	_, kbApiKeyOk := os.LookupEnv("KIBANA_API_KEY")

	if !elasticsearchEndpointsOk {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must be set for acceptance tests to run")
	}

	if !kibanaEndpointOk {
		t.Fatal("KIBANA_ENDPOINT must be set for acceptance tests to run")
	}

	authOk := (userOk && passOk) || (kbUserOk && kbPassOk) || apiKeyOk || kbApiKeyOk
	if !authOk {
		t.Fatal("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD, or KIBANA_USERNAME and KIBANA_PASSWORD, or ELASTICSEARCH_API_KEY, or KIBANA_API_KEY must be set for acceptance tests to run")
	}
}
