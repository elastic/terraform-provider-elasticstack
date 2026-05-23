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

package securityexceptionlist

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateExceptionList(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[ExceptionListModel]) (entitycore.KibanaWriteResult[ExceptionListModel], diag.Diagnostics) {
	m := req.Plan
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return entitycore.KibanaWriteResult[ExceptionListModel]{}, diags
	}

	body, d := m.toUpdateRequest(ctx, req.WriteID)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[ExceptionListModel]{}, diags
	}

	updateResp, d := kibanaoapi.UpdateExceptionList(ctx, oapiClient, req.SpaceID, *body)
	diags.Append(d...)
	if diags.HasError() {
		return entitycore.KibanaWriteResult[ExceptionListModel]{}, diags
	}

	if updateResp == nil {
		diags.AddError("Failed to update exception list", "API returned empty response")
		return entitycore.KibanaWriteResult[ExceptionListModel]{}, diags
	}

	return entitycore.KibanaWriteResult[ExceptionListModel]{Model: m}, diags
}
