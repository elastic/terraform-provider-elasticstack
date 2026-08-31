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

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var updateExceptionList = entitycore.SimpleKibanaUpdate[ExceptionListModel, kbapi.UpdateExceptionListJSONRequestBody, kbapi.SecurityExceptionsAPIExceptionList](
	func(plan ExceptionListModel, ctx context.Context, writeID string) (kbapi.UpdateExceptionListJSONRequestBody, diag.Diagnostics) {
		body, diags := plan.toUpdateRequest(ctx, writeID)
		if diags.HasError() {
			return kbapi.UpdateExceptionListJSONRequestBody{}, diags
		}
		return *body, diags
	},
	// UpdateExceptionList takes the resource ID via the request body (see
	// ExceptionListModel.toUpdateRequest), not as a separate parameter, so
	// the writeID argument required by SimpleKibanaUpdate's apiUpdate shape
	// is unused here.
	func(ctx context.Context, client *kibanaoapi.Client, spaceID, _ string, body kbapi.UpdateExceptionListJSONRequestBody) (*kbapi.SecurityExceptionsAPIExceptionList, diag.Diagnostics) {
		return kibanaoapi.UpdateExceptionList(ctx, client, spaceID, body)
	},
	(*ExceptionListModel).populateUpdated,
)

// populateUpdated captures the namespace type reported by the update
// response; NamespaceType may be defaulted by the API when the request omits
// it, so the response value is authoritative.
func (m *ExceptionListModel) populateUpdated(_ context.Context, _ string, updated *kbapi.SecurityExceptionsAPIExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics
	if entitycore.RequireNonNilKibanaWriteResponse(&diags, updated, "update", "exception list") {
		return diags
	}

	if updated.NamespaceType != "" {
		m.NamespaceType = types.StringValue(string(updated.NamespaceType))
	}

	return diags
}
