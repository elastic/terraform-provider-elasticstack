package acctest

import (
	"os"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var Providers map[string]func() (*schema.Provider, error)
var Provider *schema.Provider

func init() {
	Provider = provider.New("dev")()
	Providers = map[string]func() (*schema.Provider, error){
		"elasticstack": func() (*schema.Provider, error) {
			return Provider, nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, urlOk := os.LookupEnv("ELASTICSEARCH_URL")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")

	if !urlOk || !userOk || !passOk {
		t.Fatal("ELASTICSEARCH_URL, ELASTICSEARCH_USERNAME, ELASTICSEARCH_PASSWORD must be set for acceptance tests to run")
	}
}
