package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMuxServer(t *testing.T) {
	const providerConfig = `
	provider "elasticstack" {
		elasticsearch {
		  username  = "sup"
		  password  = "dawg"
		  endpoints = ["http://localhost:9200"]
		}
	  }
	`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"elasticstack": func() (tfprotov6.ProviderServer, error) {
				version := "acceptance_test"
				server, err := ProtoV6ProviderServerFactory(context.Background(), version)
				if err != nil {
					return nil, err
				}

				return server(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(providerConfig),
			},
		},
	})
}
