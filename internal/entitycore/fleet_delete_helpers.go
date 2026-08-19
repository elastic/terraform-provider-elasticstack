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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// SimpleFleetDelete returns a [KibanaDeleteFunc] that resolves the scoped
// Fleet client and delegates directly to apiDelete. Use for resources
// whose delete callback needs nothing beyond client, spaceID, and resourceID,
// for example:
//
//	Delete: entitycore.SimpleFleetDelete[proxyModel](fleetclient.DeleteProxy),
func SimpleFleetDelete[T KibanaResourceModel](
	apiDelete func(ctx context.Context, client *fleetclient.Client, spaceID, resourceID string) diag.Diagnostics,
) KibanaDeleteFunc[T] {
	return func(ctx context.Context, client *clients.KibanaScopedClient, resourceID string, spaceID string, _ T) diag.Diagnostics {
		return apiDelete(ctx, client.GetFleetClient(), spaceID, resourceID)
	}
}
