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

package securityexceptionitem

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func deleteExceptionItem(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model ExceptionItemModel) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()

	return kibanaoapi.DeleteExceptionListItem(ctx, oapiClient, spaceID, exceptionItemDeleteParams(resourceID, model.NamespaceType))
}

// The delete endpoint defaults namespace_type to "single", so an agnostic item
// is reported as not found - and treated as already deleted - unless the item's
// own namespace_type is sent with the request.
func exceptionItemDeleteParams(itemID string, namespaceType types.String) *kbapi.DeleteExceptionListItemParams {
	id := itemID
	params := &kbapi.DeleteExceptionListItemParams{Id: &id}
	if typeutils.IsKnown(namespaceType) {
		nsType := kbapi.SecurityExceptionsAPIExceptionNamespaceType(namespaceType.ValueString())
		params.NamespaceType = &nsType
	}
	return params
}
