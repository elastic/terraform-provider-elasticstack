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

package resource

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	esclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/connector"
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

	core := connector.PopulateCoreConnectorFieldsFromAPI(ctx, resp, &diags)
	data.ServiceType = core.ServiceType
	data.Name = core.Name
	data.Description = core.Description
	data.IndexName = core.IndexName
	data.IsNative = core.IsNative
	data.Language = core.Language
	data.APIKeyID = core.APIKeyID
	data.APIKeySecretID = core.APIKeySecretID
	data.Pipeline = core.Pipeline
	data.Scheduling = core.Scheduling
	data.Features = core.Features

	var priorConfig map[string]connector.ConfigurationValueModel
	if !data.ConfigurationValues.IsNull() && typeutils.IsKnown(data.ConfigurationValues) {
		priorConfig = typeutils.MapTypeAs[connector.ConfigurationValueModel](ctx, data.ConfigurationValues, configurationValuesPath, &diags)
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
