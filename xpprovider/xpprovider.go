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

package xpprovider

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/config"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// Package xpprovider exports needed internal types and functions used by Crossplane for instantiating, interacting and
// configuring the underlying Terraform Elasticstack providers.

// XPAPIClient exports the internal type clients.APIClient of the Terraform provider
type XPAPIClient = clients.APIClient

// Configuration exports the internal type config.ProviderConfiguration of the Terraform provider.
type Configuration = config.ProviderConfiguration

// XPProviderConfiguration is an alias for [Configuration].
//
// Deprecated: prefer [Configuration]; this name remains for backward compatibility with existing consumers.
//
//nolint:revive // Intentional legacy name; stutter is acceptable for a deprecated compatibility alias.
type XPProviderConfiguration = Configuration

// XPElasticsearchConnection exports the internal type config.ElasticsearchConnection of the Terraform provider
type XPElasticsearchConnection = config.ElasticsearchConnection

// XPKibanaConnection exports the internal type config.KibanaConnection of the Terraform provider
type XPKibanaConnection = config.KibanaConnection

// XPFleetConnection exports the internal type config.FleetConnection of the Terraform provider
type XPFleetConnection = config.FleetConnection

func NewAPIClientFromFramework(ctx context.Context, cfg Configuration, version string) (*XPAPIClient, fwdiags.Diagnostics) {
	return clients.NewAPIClientFromFramework(ctx, cfg, version)
}
