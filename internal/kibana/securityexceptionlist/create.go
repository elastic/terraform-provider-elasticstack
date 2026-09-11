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

var createExceptionList = entitycore.SimpleKibanaCreate[ExceptionListModel, kbapi.CreateExceptionListJSONRequestBody, kbapi.SecurityExceptionsAPIExceptionList](
	func(plan ExceptionListModel, ctx context.Context) (kbapi.CreateExceptionListJSONRequestBody, diag.Diagnostics) {
		body, diags := plan.toCreateRequest(ctx)
		if diags.HasError() {
			return kbapi.CreateExceptionListJSONRequestBody{}, diags
		}
		return *body, diags
	},
	kibanaoapi.CreateExceptionList,
	(*ExceptionListModel).populateCreated,
)

// populateCreated captures the ID and namespace type reported by the create
// response; NamespaceType may be defaulted by the API when the request omits
// it, so the response value is authoritative.
func (m *ExceptionListModel) populateCreated(_ context.Context, spaceID string, created *kbapi.SecurityExceptionsAPIExceptionList) diag.Diagnostics {
	var diags diag.Diagnostics
	if entitycore.RequireNonNilKibanaWriteResponse(&diags, created, "create", "exception list") {
		return diags
	}

	if created.NamespaceType != "" {
		m.NamespaceType = types.StringValue(string(created.NamespaceType))
	}

	m.ID = entitycore.KibanaResourceID(spaceID, created.Id)

	return diags
}
