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

package serverhost

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteServerHost(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model serverHostModel) diag.Diagnostics {
	var diags diag.Diagnostics
	fleetClient := client.GetFleetClient()

	// Kibana refuses to delete a Fleet server host that is currently marked
	// as default ("Default Fleet Server hosts <id> cannot be deleted"). When
	// state has default=true, clear the flag via update before deleting so
	// that `terraform destroy` succeeds without manual intervention.
	if model.Default.ValueBool() {
		isDefault := false
		_, d := fleet.UpdateFleetServerHost(ctx, fleetClient, resourceID, spaceID, kbapi.PutFleetFleetServerHostsItemidJSONRequestBody{
			IsDefault: &isDefault,
		})
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
	}

	d := fleet.DeleteFleetServerHost(ctx, fleetClient, resourceID, spaceID)
	diags.Append(d...)

	return diags
}
