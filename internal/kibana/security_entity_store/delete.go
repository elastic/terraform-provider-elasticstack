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
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteEntityStore(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	_ string,
	spaceID string,
	model tfModel,
) diag.Diagnostics {
	entityTypes, diags := expandEntityTypes(ctx, model.EntityTypes)
	if diags.HasError() {
		return diags
	}

	body := kbapi.PostSecurityEntityStoreUninstallJSONRequestBody{}
	if len(entityTypes) > 0 {
		body.EntityTypes = stringSliceToAPITypes[kbapi.PostSecurityEntityStoreUninstallJSONBodyEntityTypes](entityTypes)
	}

	diags = kibanaoapi.UninstallSecurityEntityStore(ctx, client.GetKibanaOapiClient(), spaceID, body)
	if diags.HasError() {
		return diags
	}

	diags.Append(waitForUninstall(ctx, client, spaceID)...)
	return diags
}
