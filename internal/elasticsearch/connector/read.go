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

package connector

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func readConnector(
	ctx context.Context,
	client *clients.ElasticsearchScopedClient,
	resourceID string,
	data ContentConnectorData,
) (ContentConnectorData, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	resp, getDiags := esclient.GetConnector(ctx, client, resourceID)
	diags.Append(getDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	if resp == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Connector "%s" not found, removing from state`, resourceID))
		return data, false, diags
	}

	if resp.ServiceType != nil {
		data.ServiceType = fwtypes.StringValue(*resp.ServiceType)
	}
	if resp.Name != nil {
		data.Name = fwtypes.StringValue(*resp.Name)
	}
	if resp.Description != nil {
		data.Description = fwtypes.StringValue(*resp.Description)
	}
	if resp.IndexName != nil {
		data.IndexName = fwtypes.StringValue(*resp.IndexName)
	}
	data.IsNative = fwtypes.BoolValue(resp.IsNative)
	if resp.Language != nil {
		data.Language = fwtypes.StringValue(*resp.Language)
	}
	if resp.ApiKeyId != nil {
		data.APIKeyID = fwtypes.StringValue(*resp.ApiKeyId)
	}
	if resp.ApiKeySecretId != nil {
		data.APIKeySecretID = fwtypes.StringValue(*resp.ApiKeySecretId)
	}

	data.Pipeline = populatePipelineFromAPI(ctx, resp.Pipeline, &diags)
	data.Scheduling = populateSchedulingFromAPI(ctx, resp.Scheduling, &diags)
	data.Features = populateFeaturesFromAPI(ctx, resp.Features, &diags)

	var priorConfig map[string]ConfigurationValueModel
	if !data.ConfigurationValues.IsNull() && typeutils.IsKnown(data.ConfigurationValues) {
		priorConfig = typeutils.MapTypeAs[ConfigurationValueModel](ctx, data.ConfigurationValues, configurationValuesPath, &diags)
	}
	data.ConfigurationValues = populateConfigurationValuesFromAPI(ctx, resp, priorConfig, &diags)

	id, idDiags := client.ID(ctx, resourceID)
	diags.Append(idDiags...)
	if diags.HasError() {
		return data, false, diags
	}

	data.ID = fwtypes.StringValue(id.String())
	data.ConnectorID = fwtypes.StringValue(resourceID)

	return data, true, diags
}
