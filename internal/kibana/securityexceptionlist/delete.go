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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteExceptionList(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, m ExceptionListModel) diag.Diagnostics {
	return entitycore.DeleteWithOapiClient(ctx, client, spaceID, resourceID,
		func(ctx context.Context, c *kibanaoapi.Client, spaceID, id string) diag.Diagnostics {
			params := &kbapi.DeleteExceptionListParams{Id: &id}
			if m.NamespaceType.ValueString() != "" {
				nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(m.NamespaceType.ValueString())
				params.NamespaceType = &nsType
			}
			return kibanaoapi.DeleteExceptionList(ctx, c, spaceID, params)
		})
}
