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
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteExceptionList(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, m ExceptionListModel) diag.Diagnostics {
	var diags diag.Diagnostics

	oapiClient, err := client.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("Failed to get Kibana client", err.Error())
		return diags
	}

	id := resourceID
	params := &kbapi.DeleteExceptionListParams{
		Id: &id,
	}

	if m.NamespaceType.ValueString() != "" {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
		params.NamespaceType = &nsType
	}

	diags.Append(kibanaoapi.DeleteExceptionList(ctx, oapiClient, spaceID, params)...)
	return diags
}
