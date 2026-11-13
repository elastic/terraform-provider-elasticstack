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

package advancedsettings

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type spaceGetter func(ctx context.Context, spaceID string) (*kbapi.SpaceResponse, diag.Diagnostics)

func spacesAPIGetter(client *clients.KibanaScopedClient) spaceGetter {
	return func(ctx context.Context, spaceID string) (*kbapi.SpaceResponse, diag.Diagnostics) {
		return kibanaoapi.GetSpace(ctx, client.GetKibanaOapiClient(), spaceID)
	}
}

func spaceIsMissing(ctx context.Context, spaceID string, getSpace spaceGetter) (bool, diag.Diagnostics) {
	if spaceID == clients.DefaultSpaceID {
		return false, nil
	}

	space, diags := getSpace(ctx, spaceID)
	if diags.HasError() {
		return false, diags
	}
	return space == nil, diags
}

func ensureSpaceExists(ctx context.Context, spaceID string, getSpace spaceGetter) diag.Diagnostics {
	missing, diags := spaceIsMissing(ctx, spaceID, getSpace)
	if missing {
		diags.AddError(
			"Kibana space does not exist",
			fmt.Sprintf("The space %q must exist before its advanced settings can be managed. Create the space first, for example with the elasticstack_kibana_space resource.", spaceID),
		)
	}
	return diags
}
