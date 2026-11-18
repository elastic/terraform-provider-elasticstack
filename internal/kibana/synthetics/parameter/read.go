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

package parameter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var getParameterAPI = func(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string) (*kbapi.GetParameterResponse, error) {
	return client.GetKibanaOapiClient().API.GetParameterWithResponse(ctx, resourceID, kibanautil.SpaceAwarePathRequestEditor(spaceID))
}

func readParameter(ctx context.Context, client *clients.KibanaScopedClient, resourceID, spaceID string, model Model) (Model, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	getResult, err := getParameterAPI(ctx, client, resourceID, spaceID)
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to get parameter `%s`", resourceID), err.Error())
		return model, false, diags
	}

	if getResult.StatusCode() == http.StatusNotFound {
		return model, false, diags
	}

	unwrapped, unwrapDiags := diagutil.UnwrapJSON200(getResult.JSON200, "synthetics parameter")
	diags.Append(unwrapDiags...)
	if diags.HasError() {
		return model, false, diags
	}

	result := modelFromOAPI(*unwrapped, spaceID)
	result.KibanaConnection = model.KibanaConnection

	return result, true, diags
}
