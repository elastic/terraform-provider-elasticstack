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
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func createParameter(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[Model]) (entitycore.KibanaWriteResult[Model], diag.Diagnostics) {
	plan := req.Plan
	var diags diag.Diagnostics

	kibanaClient := client.GetKibanaOapiClient()

	input := plan.toParameterRequest(false)

	// We shouldn't have to do this json marshalling ourselves,
	// https://github.com/oapi-codegen/oapi-codegen/issues/1620 means the generated code doesn't handle the oneOf
	// request body properly.
	inputJSON, err := json.Marshal(input)
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to marshal JSON for parameter `%s`", input.Key), err.Error())
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	createResult, err := kibanaClient.API.PostParametersWithBodyWithResponse(ctx, "application/json", bytes.NewReader(inputJSON), kibanautil.SpaceAwarePathRequestEditor(req.SpaceID))
	if err != nil {
		diags.AddError(fmt.Sprintf("Failed to create parameter `%s`", input.Key), err.Error())
		return entitycore.KibanaWriteResult[Model]{}, diags
	}

	createResponse, parseDiags := parseCreateParameterResponse(createResult, input.Key)
	if parseDiags.HasError() {
		return entitycore.KibanaWriteResult[Model]{}, parseDiags
	}

	plan.setCompositeIdentity(req.SpaceID, *createResponse.Id)

	return entitycore.KibanaWriteResult[Model]{Model: plan}, diags
}
