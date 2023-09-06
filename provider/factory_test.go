package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
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
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"elasticstack": func() (tfprotov5.ProviderServer, error) {
				version := "test"
				sdkv2Provider := New(version)
				frameworkProvider := providerserver.NewProtocol5(NewFrameworkProvider(version))
				ctx := context.Background()
				providers := []func() tfprotov5.ProviderServer{
					frameworkProvider,
					sdkv2Provider.GRPCProvider,
				}

				muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

				if err != nil {
					return nil, err
				}

				return muxServer.ProviderServer(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(providerConfig),
			},
		},
	})
}
