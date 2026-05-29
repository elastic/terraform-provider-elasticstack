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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	attrKuery             = "kuery"
	attrPage              = "page"
	attrPerPage           = "per_page"
	attrCloudConnectors   = "cloud_connectors"
	defaultSpaceID        = "default"
)

type cloudConnectorsDataSourceModel struct {
	entitycore.KibanaConnectionField
	ID              types.String `tfsdk:"id"`
	SpaceID         types.String `tfsdk:"space_id"`
	Kuery           types.String `tfsdk:"kuery"`
	Page            types.Int64  `tfsdk:"page"`
	PerPage         types.Int64  `tfsdk:"per_page"`
	CloudConnectors types.List   `tfsdk:"cloud_connectors"`
}

type cloudConnectorListItemModel struct {
	ID                    types.String `tfsdk:"id"`
	CloudConnectorID      types.String `tfsdk:"cloud_connector_id"`
	SpaceID               types.String `tfsdk:"space_id"`
	Name                  types.String `tfsdk:"name"`
	CloudProvider         types.String `tfsdk:"cloud_provider"`
	AccountType           types.String `tfsdk:"account_type"`
	Namespace             types.String `tfsdk:"namespace"`
	PackagePolicyCount    types.Int64  `tfsdk:"package_policy_count"`
	VerificationStatus    types.String `tfsdk:"verification_status"`
	VerificationStartedAt types.String `tfsdk:"verification_started_at"`
	VerificationFailedAt  types.String `tfsdk:"verification_failed_at"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

func (m cloudConnectorsDataSourceModel) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{
		{
			MinVersion:   *cloudConnectorMinVersion,
			ErrorMessage: fmt.Sprintf("Fleet cloud connectors require Kibana v%s or later.", cloudConnectorMinVersion),
		},
	}, nil
}

func mapAPIToDatasourceModel(ctx context.Context, model *cloudConnectorsDataSourceModel, spaceID string, items []fleetclient.CloudConnectorItem) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ID = types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: "cloud_connectors"}).String())
	model.CloudConnectors = typeutils.SliceToListType(ctx, items, getCloudConnectorListItemElemType(ctx), path.Root(attrCloudConnectors), &diags,
		func(item fleetclient.CloudConnectorItem, _ typeutils.ListMeta) cloudConnectorListItemModel {
			return mapAPIItemToListItem(spaceID, item)
		})

	return diags
}

func mapAPIItemToListItem(spaceID string, item fleetclient.CloudConnectorItem) cloudConnectorListItemModel {
	model := cloudConnectorListItemModel{
		ID:               types.StringValue((&clients.CompositeID{ClusterID: spaceID, ResourceID: item.ID}).String()),
		CloudConnectorID: types.StringValue(item.ID),
		SpaceID:          types.StringValue(spaceID),
		Name:             types.StringValue(item.Name),
		CloudProvider:    types.StringValue(item.CloudProvider),
	}

	if item.AccountType != nil && *item.AccountType != "" {
		model.AccountType = types.StringValue(*item.AccountType)
	} else {
		model.AccountType = types.StringNull()
	}

	if item.Namespace != nil && *item.Namespace != "" {
		model.Namespace = types.StringValue(*item.Namespace)
	} else {
		model.Namespace = types.StringNull()
	}

	model.PackagePolicyCount = types.Int64Value(int64(item.PackagePolicyCount))

	if item.VerificationStatus != nil && *item.VerificationStatus != "" {
		model.VerificationStatus = types.StringValue(*item.VerificationStatus)
	} else {
		model.VerificationStatus = types.StringNull()
	}

	if item.VerificationStartedAt != nil && *item.VerificationStartedAt != "" {
		model.VerificationStartedAt = types.StringValue(*item.VerificationStartedAt)
	} else {
		model.VerificationStartedAt = types.StringNull()
	}

	if item.VerificationFailedAt != nil && *item.VerificationFailedAt != "" {
		model.VerificationFailedAt = types.StringValue(*item.VerificationFailedAt)
	} else {
		model.VerificationFailedAt = types.StringNull()
	}

	if item.CreatedAt != "" {
		model.CreatedAt = types.StringValue(item.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}

	if item.UpdatedAt != "" {
		model.UpdatedAt = types.StringValue(item.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}

	return model
}

var (
	_ entitycore.WithVersionRequirements = (*cloudConnectorsDataSourceModel)(nil)
)
