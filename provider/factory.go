package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProtoV5ProviderServerFactory returns a muxed terraform-plugin-go protocol v5 provider factory function.
// Plugin SDK V2 provider server is also returned (useful for testing).
func ProtoV5ProviderServerFactory(ctx context.Context, version string) (func() tfprotov5.ProviderServer, *schema.Provider, error) {
	sdkv2Provider := New(version)

	servers := []func() tfprotov5.ProviderServer{
		sdkv2Provider.GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, servers...)
	if err != nil {
		return nil, nil, err
	}

	return muxServer.ProviderServer, sdkv2Provider, nil
}
