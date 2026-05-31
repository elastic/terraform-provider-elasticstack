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

package integrationds

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readDataSource(
	ctx context.Context,
	kbClient *clients.KibanaScopedClient,
	resourceID, spaceID string,
	config integrationDataSourceModel,
) (integrationDataSourceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	client := kbClient.GetFleetClient()

	prerelease := config.Prerelease.ValueBool()
	packages, pDiags := fleet.GetPackages(ctx, client, prerelease, spaceID)
	diags.Append(pDiags...)
	if diags.HasError() {
		return config, false, diags
	}

	(&config).populateFromAPI(resourceID, packages)

	if config.Version.IsNull() {
		return config, false, diags
	}

	hash, err := typeutils.StringToHash(resourceID)
	if err != nil {
		diags.AddError(err.Error(), "")
		return config, false, diags
	}
	config.ID = types.StringPointerValue(hash)
	config.Name = types.StringValue(resourceID)
	config.SpaceID = types.StringValue(spaceID)

	return config, true, diags
}
