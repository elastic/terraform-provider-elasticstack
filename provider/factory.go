// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
		ctx,
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
