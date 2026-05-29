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

package cloudconnector

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resolveWireVarsSecrets(
	ctx context.Context,
	es *elasticsearch.TypedClient,
	cloudProvider string,
	vars postWireVars,
) (postWireVars, diag.Diagnostics) {
	if es == nil || len(vars) == 0 {
		return vars, nil
	}

	var diags diag.Diagnostics
	out := make(postWireVars, len(vars))
	for key, wireVar := range vars {
		resolved, resolveDiags := resolveWireVarSecret(ctx, es, cloudProvider, key, wireVar)
		diags.Append(resolveDiags...)
		if diags.HasError() {
			return nil, diags
		}
		out[key] = resolved
	}
	return out, diags
}

func resolveWireVarSecret(
	ctx context.Context,
	es *elasticsearch.TypedClient,
	cloudProvider, key string,
	wireVar kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties,
) (kbapi.PostFleetCloudConnectorsJSONBody_Vars_AdditionalProperties, diag.Diagnostics) {
	if !requiresFleetSecretRef(cloudProvider, key) {
		return wireVar, nil
	}

	structured, err := wireVar.AsPostFleetCloudConnectorsJSONBodyVars3()
	if err != nil || structured.Type != varsStructuredTypePassword {
		return wireVar, nil
	}

	if _, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value1(); err == nil {
		return wireVar, nil
	}

	plainValue, err := structured.Value.AsPostFleetCloudConnectorsJSONBodyVars3Value0()
	if err != nil || plainValue == "" {
		return wireVar, nil
	}

	secretID, secretDiags := fleetclient.CreateFleetSecret(ctx, es, plainValue)
	if secretDiags.HasError() {
		return wireVar, secretDiags
	}

	return wireStructuredSecretRefPost(varsStructuredTypePassword, cloudConnectorSecretRef{
		ID:          types.StringValue(secretID),
		IsSecretRef: types.BoolValue(true),
	})
}

func requiresFleetSecretRef(cloudProvider, key string) bool {
	switch cloudProvider {
	case cloudProviderAWS:
		return key == attrAWSExternalID
	case cloudProviderAzure:
		return key == attrAzureTenantID || key == attrAzureClientID
	default:
		return false
	}
}
