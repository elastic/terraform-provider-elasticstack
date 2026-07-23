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
	"fmt"

	kboapi "github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func parseCreateParameterResponse(resp *kboapi.PostParametersResponse, key string) (kboapi.SyntheticsPostParameterResponse, diag.Diagnostics) {
	var empty kboapi.SyntheticsPostParameterResponse
	if resp == nil {
		return empty, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				fmt.Sprintf("Failed to create parameter `%s`", key),
				"API returned no response",
			),
		}
	}

	_, diags := kibanaoapi.HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *kboapi.CreateParamResponse {
			return resp.JSON200
		})
	if diags.HasError() {
		return empty, diags
	}

	createResponse, err := resp.JSON200.AsSyntheticsPostParameterResponse()
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to parse parameter response `%s`", key), err.Error())
		return empty, diags
	}

	if createResponse.Id == nil {
		diags.AddError(fmt.Sprintf("Unexpected nil id in create parameter response `%s`", key), "")
		return empty, diags
	}

	return createResponse, diags
}
