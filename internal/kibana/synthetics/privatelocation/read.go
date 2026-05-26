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

package privatelocation

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func readPrivateLocation(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oapiClient := client.GetKibanaOapiClient()

	result, dg := kibanaoapi.GetPrivateLocation(ctx, oapiClient, spaceID, resourceID)
	diags.Append(dg...)
	if diags.HasError() {
		return model, false, diags
	}

	// nil result means HTTP 404 — resource no longer exists.
	if result == nil {
		return model, false, diags
	}

	readModel := modelFromAPI(*result, spaceID)
	readModel.KibanaConnection = model.KibanaConnection
	readModel.Geo = preserveGeoFromInput(ctx, model.Geo, readModel.Geo)

	return readModel, true, diags
}

// preserveGeoFromInput keeps the input geo when it is semantically equal to the
// API value under float32 precision; otherwise the API value is used.
func preserveGeoFromInput(ctx context.Context, input, api *tfGeoConfigV0) *tfGeoConfigV0 {
	if input == nil || api == nil {
		return api
	}

	latEqual, _ := input.Lat.Float64SemanticEquals(ctx, api.Lat)
	lonEqual, _ := input.Lon.Float64SemanticEquals(ctx, api.Lon)
	if latEqual && lonEqual {
		return input
	}
	return api
}
