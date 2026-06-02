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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func updateEntityStore(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan
	prior := *req.Prior
	spaceID := normalizeSpaceID(plan.SpaceID)

	added, removed, diags := diffEntityTypes(ctx, prior.EntityTypes, plan.EntityTypes)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	if !plan.LogExtraction.Equal(prior.LogExtraction) {
		body, d := buildUpdateBody(ctx, plan)
		if d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
		if d := kibanaoapi.UpdateSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, body); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	}

	if len(added) > 0 {
		body, d := buildInstallBody(ctx, plan)
		if d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
		if d := kibanaoapi.InstallSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, body); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	}

	if len(removed) > 0 {
		if !typeutils.IsKnown(plan.AllowEntityTypeShrink) || !plan.AllowEntityTypeShrink.ValueBool() {
			var out diag.Diagnostics
			out.AddError("Entity type shrink blocked", "Removing values from entity_types requires allow_entity_type_shrink = true. No API calls were made.")
			return entitycore.KibanaWriteResult[tfModel]{}, out
		}
		if d := kibanaoapi.UninstallSecurityEntityStore(
			ctx,
			client.GetKibanaOapiClient(),
			spaceID,
			kbapi.PostSecurityEntityStoreUninstallJSONRequestBody{EntityTypes: stringSliceToAPITypes[kbapi.PostSecurityEntityStoreUninstallJSONBodyEntityTypes](removed)},
		); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	}

	if !plan.Started.Equal(prior.Started) {
		if plan.Started.ValueBool() {
			if d := kibanaoapi.StartSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, kbapi.PutSecurityEntityStoreStartJSONRequestBody{}); d.HasError() {
				return entitycore.KibanaWriteResult[tfModel]{}, d
			}
		} else {
			if d := kibanaoapi.StopSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, kbapi.PutSecurityEntityStoreStopJSONRequestBody{}); d.HasError() {
				return entitycore.KibanaWriteResult[tfModel]{}, d
			}
		}
	}

	plan.SpaceID = types.StringValue(spaceID)
	plan.ID = types.StringValue(buildID(spaceID))
	return entitycore.KibanaWriteResult[tfModel]{Model: plan}, nil
}
