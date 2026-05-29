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

	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAPIToDatasourceModel_emptyList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var model cloudConnectorsDataSourceModel

	diags := mapAPIToDatasourceModel(ctx, &model, "default", []fleetclient.CloudConnectorItem{})
	require.False(t, diags.HasError())

	assert.Equal(t, "default/cloud_connectors", model.ID.ValueString())
	require.False(t, model.CloudConnectors.IsNull())
	require.False(t, model.CloudConnectors.IsUnknown())
	assert.Empty(t, model.CloudConnectors.Elements())
}

func TestMapAPIToDatasourceModel_multipleItems(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	accountType := "single-account"
	namespace := "default"
	verificationStatus := "verified"
	verificationStartedAt := "2026-01-01T00:00:00.000Z"
	items := []fleetclient.CloudConnectorItem{
		{
			ID:                 "conn-aws-1",
			Name:               "aws-connector",
			CloudProvider:      "aws",
			AccountType:        &accountType,
			Namespace:          &namespace,
			PackagePolicyCount: 2,
			Vars: map[string]any{
				"role_arn": "arn:aws:iam::123456789012:role/Elastic",
			},
			VerificationStatus:    &verificationStatus,
			VerificationStartedAt: &verificationStartedAt,
			CreatedAt:             "2026-01-01T00:00:00.000Z",
			UpdatedAt:             "2026-01-02T00:00:00.000Z",
		},
		{
			ID:                 "conn-azure-1",
			Name:               "azure-connector",
			CloudProvider:      "azure",
			PackagePolicyCount: 0,
			CreatedAt:          "2026-01-03T00:00:00.000Z",
			UpdatedAt:          "2026-01-04T00:00:00.000Z",
		},
	}

	var model cloudConnectorsDataSourceModel
	diags := mapAPIToDatasourceModel(ctx, &model, "custom-space", items)
	require.False(t, diags.HasError())

	require.Len(t, model.CloudConnectors.Elements(), 2)

	firstObj, ok := model.CloudConnectors.Elements()[0].(types.Object)
	require.True(t, ok)
	first := objectToListItemModel(t, firstObj)

	assert.Equal(t, "custom-space/conn-aws-1", first.ID.ValueString())
	assert.Equal(t, "conn-aws-1", first.CloudConnectorID.ValueString())
	assert.Equal(t, "custom-space", first.SpaceID.ValueString())
	assert.Equal(t, "aws-connector", first.Name.ValueString())
	assert.Equal(t, "aws", first.CloudProvider.ValueString())
	assert.Equal(t, "single-account", first.AccountType.ValueString())
	assert.Equal(t, "default", first.Namespace.ValueString())
	assert.Equal(t, int64(2), first.PackagePolicyCount.ValueInt64())
	assert.Equal(t, "verified", first.VerificationStatus.ValueString())
	assert.Equal(t, "2026-01-01T00:00:00.000Z", first.VerificationStartedAt.ValueString())
	assert.True(t, first.VerificationFailedAt.IsNull())
	assert.Equal(t, "2026-01-01T00:00:00.000Z", first.CreatedAt.ValueString())
	assert.Equal(t, "2026-01-02T00:00:00.000Z", first.UpdatedAt.ValueString())

	secondObj, ok := model.CloudConnectors.Elements()[1].(types.Object)
	require.True(t, ok)
	second := objectToListItemModel(t, secondObj)

	assert.Equal(t, "azure-connector", second.Name.ValueString())
	assert.True(t, second.AccountType.IsNull())
	assert.True(t, second.Namespace.IsNull())
	assert.True(t, second.VerificationStatus.IsNull())
	assert.True(t, second.VerificationStartedAt.IsNull())
	assert.True(t, second.VerificationFailedAt.IsNull())
}

func TestMapAPIItemToListItem_noVarsInModel(t *testing.T) {
	t.Parallel()

	item := fleetclient.CloudConnectorItem{
		ID:            "conn-1",
		Name:          "connector",
		CloudProvider: "aws",
		Vars: map[string]any{
			"role_arn": "secret-ish",
		},
		CreatedAt: "2026-01-01T00:00:00.000Z",
		UpdatedAt: "2026-01-02T00:00:00.000Z",
	}

	model := mapAPIItemToListItem("default", item)
	obj, diags := types.ObjectValue(getCloudConnectorListItemAttrTypes(), map[string]attr.Value{
		attrID:                    model.ID,
		attrCloudConnectorID:      model.CloudConnectorID,
		attrSpaceID:               model.SpaceID,
		attrName:                  model.Name,
		attrCloudProvider:         model.CloudProvider,
		attrAccountType:           model.AccountType,
		attrNamespace:             model.Namespace,
		attrPackagePolicyCount:    model.PackagePolicyCount,
		attrVerificationStatus:    model.VerificationStatus,
		attrVerificationStartedAt: model.VerificationStartedAt,
		attrVerificationFailedAt:  model.VerificationFailedAt,
		attrCreatedAt:             model.CreatedAt,
		attrUpdatedAt:             model.UpdatedAt,
	})
	require.False(t, diags.HasError())

	attrs := obj.Attributes()
	for _, excluded := range []string{attrVarsMap, attrAWSBlock, attrAzureBlock} {
		_, ok := attrs[excluded]
		assert.False(t, ok, "list item must not contain %q", excluded)
	}
}

func objectToListItemModel(t *testing.T, obj types.Object) cloudConnectorListItemModel {
	t.Helper()

	attrs := obj.Attributes()
	return cloudConnectorListItemModel{
		ID:                    attrs[attrID].(types.String),
		CloudConnectorID:      attrs[attrCloudConnectorID].(types.String),
		SpaceID:               attrs[attrSpaceID].(types.String),
		Name:                  attrs[attrName].(types.String),
		CloudProvider:         attrs[attrCloudProvider].(types.String),
		AccountType:           attrs[attrAccountType].(types.String),
		Namespace:             attrs[attrNamespace].(types.String),
		PackagePolicyCount:    attrs[attrPackagePolicyCount].(types.Int64),
		VerificationStatus:    attrs[attrVerificationStatus].(types.String),
		VerificationStartedAt: attrs[attrVerificationStartedAt].(types.String),
		VerificationFailedAt:  attrs[attrVerificationFailedAt].(types.String),
		CreatedAt:             attrs[attrCreatedAt].(types.String),
		UpdatedAt:             attrs[attrUpdatedAt].(types.String),
	}
}

func getCloudConnectorListItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		attrID:                    types.StringType,
		attrCloudConnectorID:      types.StringType,
		attrSpaceID:               types.StringType,
		attrName:                  types.StringType,
		attrCloudProvider:         types.StringType,
		attrAccountType:           types.StringType,
		attrNamespace:             types.StringType,
		attrPackagePolicyCount:    types.Int64Type,
		attrVerificationStatus:    types.StringType,
		attrVerificationStartedAt: types.StringType,
		attrVerificationFailedAt:  types.StringType,
		attrCreatedAt:             types.StringType,
		attrUpdatedAt:             types.StringType,
	}
}
