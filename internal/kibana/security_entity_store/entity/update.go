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

package entity

import (
	"bytes"
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateEntity(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	req entitycore.KibanaWriteRequest[tfModel],
) (entitycore.KibanaWriteResult[tfModel], diag.Diagnostics) {
	plan := req.Plan

	spaceID, entityType, bodyBytes, diags := buildEntityWriteBody(ctx, plan)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, diags
	}

	force := false
	if !plan.Force.IsNull() && !plan.Force.IsUnknown() {
		force = plan.Force.ValueBool()
	}

	if d := kibanaoapi.UpdateSecurityEntityStoreEntity(ctx, client.GetKibanaOapiClient(), spaceID, entityType, bytes.NewReader(bodyBytes), force); d.HasError() {
		return entitycore.KibanaWriteResult[tfModel]{}, d
	}

	return entitycore.KibanaWriteResult[tfModel]{Model: plan}, nil
}
