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
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPlanVarsMapFromAWSBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	awsObj, diags := types.ObjectValue(awsAttrTypes(), map[string]attr.Value{
		attrAWSRoleArn:             types.StringValue("arn:aws:iam::123:role/x"),
		attrAWSExternalID:          types.StringValue("secret"),
		attrAWSExternalIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	})
	require.False(t, diags.HasError())

	config := cloudConnectorModel{
		AWS: awsObj,
	}

	varsMap, mapDiags := planVarsMapFromAWSBlock(ctx, config)
	require.False(t, mapDiags.HasError(), mapDiags)
	require.False(t, varsMap.IsNull())
	require.False(t, varsMap.IsUnknown())
}

func TestPlanVarsMapFromAzureBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	azureObj, diags := types.ObjectValue(azureAttrTypes(), map[string]attr.Value{
		attrAzureCloudConnectorID:  types.StringValue("azure-conn"),
		attrAzureTenantID:          types.StringValue("tenant"),
		attrAzureClientID:          types.StringValue("client"),
		attrAzureTenantIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
		attrAzureClientIDSecretRef: types.ObjectNull(secretRefAttrTypes()),
	})
	require.False(t, diags.HasError())

	config := cloudConnectorModel{
		Azure: azureObj,
	}

	varsMap, mapDiags := planVarsMapFromAzureBlock(ctx, config)
	require.False(t, mapDiags.HasError(), mapDiags)
	require.False(t, varsMap.IsNull())
	require.False(t, varsMap.IsUnknown())
}
