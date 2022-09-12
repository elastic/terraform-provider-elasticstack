package provider_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
)

func TestProvider(t *testing.T) {
	if err := acctest.Provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
