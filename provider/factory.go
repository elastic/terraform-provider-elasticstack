package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

// ProtoV5ProviderServerFactory returns a muxed terraform-plugin-go protocol v5 provider factory function.
func ProtoV5ProviderServerFactory(ctx context.Context, version string) (func() tfprotov5.ProviderServer, error) {
	sdkv2Provider := New(version)
	frameworkProvider := providerserver.NewProtocol5(NewFrameworkProvider(version))

	servers := []func() tfprotov5.ProviderServer{
		frameworkProvider,
		sdkv2Provider.GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, servers...)
	if err != nil {
		return nil, fmt.Errorf("initialize mux server: %w", err)
	}

	return muxServer.ProviderServer, nil
}
