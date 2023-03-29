package acctest

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

var Providers map[string]func() (tfprotov5.ProviderServer, error)

func init() {
	providerServerFactory, err := provider.ProtoV5ProviderServerFactory(context.Background(), "dev")
	if err != nil {
		log.Fatal(err)
	}
	Providers = map[string]func() (tfprotov5.ProviderServer, error){
		"elasticstack": func() (tfprotov5.ProviderServer, error) {
			return providerServerFactory(), nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, elasticsearchEndpointsOk := os.LookupEnv("ELASTICSEARCH_ENDPOINTS")
	_, kibanaEndpointOk := os.LookupEnv("KIBANA_ENDPOINT")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")

	if !elasticsearchEndpointsOk {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must be set for acceptance tests to run")
	}

	if !kibanaEndpointOk {
		t.Fatal("KIBANA_ENDPOINT must be set for acceptance tests to run")
	}

	// Technically ES tests can use the API Key, however username/password is required for Kibana tests.
	usernamePasswordOk := userOk && passOk
	if !usernamePasswordOk {
		t.Fatal("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set for acceptance tests to run")
	}
}
