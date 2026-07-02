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

package security_entity_store_entity_link

import (
	"context"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// createRetryPollInterval is the cadence at which the entity-link create call
// is retried while the entity store is still initializing (HTTP 500). The
// overall retry budget is bounded by the Create ctx deadline (from the resource
// timeouts block), not by this interval.
const createRetryPollInterval = 5 * time.Second

func createEntityLink(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[entityLinkModel]) (entitycore.KibanaWriteResult[entityLinkModel], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	entityIDs := typeutils.SetTypeAs[string](ctx, plan.EntityIDs, path.Root("entity_ids"), &diags)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, diags
	}

	body := kbapi.PostSecurityEntityStoreResolutionLinkJSONRequestBody{
		TargetId:  plan.TargetID.ValueString(),
		EntityIds: entityIDs,
	}

	attempt := func(ctx context.Context) (int, []byte, error) {
		resp, err := client.GetKibanaOapiClient().API.PostSecurityEntityStoreResolutionLinkWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(req.SpaceID))
		if err != nil {
			return 0, nil, err
		}
		return resp.StatusCode(), resp.Body, nil
	}

	if d := kibanaoapi.RetryCreateOnServerError(ctx, "security entity store entity link", plan.TargetID.ValueString(), attempt, createRetryPollInterval); d.HasError() {
		return entitycore.KibanaWriteResult[entityLinkModel]{}, d
	}

	return entitycore.KibanaWriteResult[entityLinkModel]{Model: plan}, diags
}
