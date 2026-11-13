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

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func deleteAdvancedSettings(ctx context.Context, client *clients.KibanaScopedClient, _, spaceID string, model advancedSettingsModel) diag.Diagnostics {
	missing, diags := spaceIsMissing(ctx, spaceID, spacesAPIGetter(client))
	if missing || diags.HasError() {
		return diags
	}

	diags.Append(applyChanges(ctx, client, spaceID, resetChanges(model), false)...)
	return diags
}

func resetChanges(model advancedSettingsModel) map[string]any {
	changes := map[string]any{}
	for _, managed := range []types.Map{model.StringSettings, model.JSONSettings} {
		for key := range managed.Elements() {
			changes[key] = nil
		}
	}
	return changes
}
