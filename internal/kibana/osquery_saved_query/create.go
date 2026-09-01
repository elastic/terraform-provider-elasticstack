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

package osquerysavedquery

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func createOsquerySavedQuery(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[osquerySavedQueryModel],
) (entitycore.KibanaWriteResult[osquerySavedQueryModel], diag.Diagnostics) {
	return entitycore.SimpleKibanaCreate[osquerySavedQueryModel, kbapi.OsqueryCreateSavedQueryJSONRequestBody, kibanaoapi.OsquerySavedQueryCreateEntity](
		func(plan osquerySavedQueryModel, ctx context.Context) (kbapi.OsqueryCreateSavedQueryJSONRequestBody, diag.Diagnostics) {
			return plan.toAPICreateRequest(ctx)
		},
		kibanaoapi.CreateOsquerySavedQuery,
		func(plan *osquerySavedQueryModel, ctx context.Context, spaceID string, entity *kibanaoapi.OsquerySavedQueryCreateEntity) diag.Diagnostics {
			var diags diag.Diagnostics
			diags.Append(prebuiltGuardDiagnostic(entity.Prebuilt)...)
			if diags.HasError() {
				return diags
			}
			diags.Append(plan.populateFromCreateAPI(ctx, entity)...)
			if diags.HasError() {
				return diags
			}
			plan.SpaceID = types.StringValue(spaceID)
			return diags
		},
	)(ctx, client, req)
}
