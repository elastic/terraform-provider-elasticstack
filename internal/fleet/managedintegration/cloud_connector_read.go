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

package managedintegration

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// applyCloudConnectorFromAPI merges cloud_connector.enabled and
// cloud_connector_id from the managed_integrations GET response while
// preserving write-only name and target_csp from prior state when present.
// When prior had no block (import), the block is built from API values with
// null write-only fields.
func (m *managedIntegrationModel) applyCloudConnectorFromAPI(ctx context.Context, item *kbapi.KibanaHTTPAPIsManagedIntegration, diags *diag.Diagnostics) {
	if item == nil {
		return
	}

	var preserveName, preserveTarget types.String
	hadPriorBlock := typeutils.IsKnown(m.CloudConnector) && !m.CloudConnector.IsNull()
	if hadPriorBlock {
		var prior cloudConnectorModel
		diags.Append(m.CloudConnector.As(ctx, &prior, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return
		}
		preserveName = prior.Name
		preserveTarget = prior.TargetCSP
	}

	if item.CloudConnector == nil {
		if hadPriorBlock {
			m.CloudConnector = types.ObjectNull(cloudConnectorAttrTypes())
		}
		return
	}

	api := item.CloudConnector
	cc := cloudConnectorModel{
		Enabled:          types.BoolValue(api.Enabled),
		CloudConnectorID: types.StringValue(api.CloudConnectorId),
	}
	if hadPriorBlock {
		cc.Name = preserveName
		cc.TargetCSP = preserveTarget
	} else {
		cc.Name = types.StringNull()
		cc.TargetCSP = types.StringNull()
	}

	obj, d := types.ObjectValueFrom(ctx, cloudConnectorAttrTypes(), cc)
	diags.Append(d...)
	m.CloudConnector = obj
}
