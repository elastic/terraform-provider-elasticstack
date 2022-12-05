package acctest

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var Providers map[string]func() (tfprotov5.ProviderServer, error)
var Provider *schema.Provider

func init() {
	providerServerFactory, sdkv2Provider, err := provider.ProtoV5ProviderServerFactory(context.Background(), "dev")
	if err != nil {
		log.Fatal(err)
	}
	Provider = sdkv2Provider
	Providers = map[string]func() (tfprotov5.ProviderServer, error){
		"elasticstack": func() (tfprotov5.ProviderServer, error) {
			return providerServerFactory(), nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, endpointsOk := os.LookupEnv("ELASTICSEARCH_ENDPOINTS")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")
	_, apikeyOk := os.LookupEnv("ELASTICSEARCH_API_KEY")

	if !endpointsOk {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must be set for acceptance tests to run")
	}

	usernamePasswordOk := userOk && passOk
	if !((!usernamePasswordOk && apikeyOk) || (usernamePasswordOk && !apikeyOk)) {
		t.Fatal("Either ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD must be set, or ELASTICSEARCH_API_KEY must be set for acceptance tests to run")
	}
}
