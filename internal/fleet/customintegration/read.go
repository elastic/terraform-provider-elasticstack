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

package customintegration

import (
	"context"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readCustomIntegration(ctx context.Context, client *clients.KibanaScopedClient, _, spaceID string, model customIntegrationModel) (customIntegrationModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	fleetClient := client.GetFleetClient()

	pkg, pkgDiags := fleet.GetPackage(ctx, fleetClient, model.PackageName.ValueString(), model.PackageVersion.ValueString(), spaceID)
	if pkgDiags.HasError() {
		diags.Append(pkgDiags...)
		return model, false, diags
	}

	if pkg == nil {
		packages, listDiags := fleet.GetPackages(ctx, fleetClient, true, spaceID)
		if listDiags.HasError() {
			diags.Append(listDiags...)
			return model, false, diags
		}
		for _, candidate := range packages {
			if candidate.Name != model.PackageName.ValueString() || candidate.Version != model.PackageVersion.ValueString() {
				continue
			}
			if candidate.Status != nil && strings.EqualFold(*candidate.Status, "installed") {
				return model, true, nil
			}
		}
		return model, false, nil
	}

	if pkg.Status == nil || *pkg.Status != "installed" {
		return model, false, nil
	}

	return model, true, nil
}
