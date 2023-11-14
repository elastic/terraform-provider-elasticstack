package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
)

func ProtoV6ProviderServerFactory(ctx context.Context, version string) (func() tfprotov6.ProviderServer, error) {
	sdkv2Provider := New(version)
	frameworkProvider := providerserver.NewProtocol6(NewFrameworkProvider(version))

	upgradedSdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		sdkv2Provider.GRPCProvider,
	)

	if err != nil {
		return nil, fmt.Errorf("cannot upgrade the SDKv2 provider to protocol 6: %w", err)
	}

	servers := []func() tfprotov6.ProviderServer{
		frameworkProvider,
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, servers...)
	if err != nil {
		return nil, fmt.Errorf("initialize mux server: %w", err)
	}

	return muxServer.ProviderServer, nil
}
