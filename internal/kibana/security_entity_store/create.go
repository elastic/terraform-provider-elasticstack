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
	"time"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	entity "github.com/elastic/terraform-provider-elasticstack/internal/kibana/security_entity_store/entity"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// installRetryPollInterval is the cadence at which the entity store install is
// retried while the store is still initializing (HTTP 500). The overall retry
// budget is bounded by the Create ctx deadline (from the resource timeouts
// block), not by this interval.
const installRetryPollInterval = 5 * time.Second

func createEntityStore(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan
	spaceID := entity.NormalizeSpaceID(plan.SpaceID)
	body, diags := buildInstallBody(ctx, plan)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	install := func(ctx context.Context) (int, []byte, error) {
		return kibanaoapi.InstallSecurityEntityStoreStatus(ctx, client.GetKibanaOapiClient(), spaceID, body)
	}
	if d := kibanaoapi.RetryCreateOnServerError(ctx, "security entity store", spaceID, install, installRetryPollInterval); d.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, d
	}

	if !plan.Started.IsNull() && !plan.Started.IsUnknown() && !plan.Started.ValueBool() {
		if d := kibanaoapi.StopSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, kbapi.PutSecurityEntityStoreStopJSONRequestBody{}); d.HasError() {
			return entitycore.KibanaWriteResult[tfModel]{}, d
		}
	}

	plan.SpaceID = types.StringValue(spaceID)
	plan.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: resourceID}).String())
	return entitycore.KibanaWriteResult[tfModel]{Model: plan}, nil
}
