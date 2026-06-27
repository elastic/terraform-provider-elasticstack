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

package tag

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func deleteTag(
	ctx context.Context,
	client *clients.KibanaScopedClient,
	resourceID string,
	spaceID string,
	_ tagModel,
) diag.Diagnostics {
	oapiClient := client.GetKibanaOapiClient()

	existing, getDiags := getTagAPI(ctx, oapiClient, spaceID, resourceID)
	if getDiags.HasError() {
		return getDiags
	}

	if existing == nil {
		return nil
	}

	if managedDiags := checkManagedTag(existing); managedDiags.HasError() {
		return managedDiags
	}

	return deleteTagAPI(ctx, oapiClient, spaceID, resourceID)
}
