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

package security_entity_store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readEntityStoreDataSource(ctx context.Context, client *clients.KibanaScopedClient, model dsModel) (dsModel, diag.Diagnostics) {
	if supported, diags := client.EnforceMinVersion(ctx, MinVersion); diags.HasError() {
		return model, diags
	} else if !supported {
		var out diag.Diagnostics
		out.AddError("Unsupported server version", fmt.Sprintf("elasticstack_kibana_security_entity_store_status is supported only for Kibana v%s and above", MinVersion.String()))
		return model, out
	}

	spaceID := normalizeSpaceID(model.SpaceID)
	includeComponents := !model.IncludeComponents.IsNull() && !model.IncludeComponents.IsUnknown() && model.IncludeComponents.ValueBool()
	status, rawBody, diags := getEntityStoreStatus(ctx, client, spaceID, includeComponents)
	if diags.HasError() {
		return model, diags
	}

	engines := status.Engines
	if engines == nil {
		engines = []entityStoreEngine{}
	}
	enginesJSON, err := json.Marshal(engines)
	if err != nil {
		return model, diagutil.FrameworkDiagFromError(err)
	}
	statusJSON, marshalDiags := normalizeJSONBytes(rawBody)
	if marshalDiags.HasError() {
		return model, marshalDiags
	}

	model.SpaceID = types.StringValue(spaceID)
	model.Installed = types.BoolValue(string(status.Status) != "not_installed")
	model.OverallStatus = types.StringValue(string(status.Status))
	model.EnginesJSON = types.StringValue(string(enginesJSON))
	model.StatusJSON = types.StringValue(statusJSON)
	return model, nil
}
