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

package connectors

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func preconfiguredConnectorIDUnsupportedMessage() string {
	return "Preconfigured connector IDs are only supported for Elastic Stack v" +
		MinVersionSupportingPreconfiguredIDs.String() +
		" and above. Either remove the `connector_id` attribute or upgrade your target cluster to supported version"
}

// userSuppliedConnectorID reports whether the Terraform plan includes a
// user-configured connector_id (as opposed to a post-create API-assigned UUID).
func userSuppliedConnectorID(model tfModel) bool {
	return typeutils.IsKnown(model.ConnectorID) && model.ConnectorID.ValueString() != ""
}

// enforceUserSuppliedConnectorIDVersion gates create when the user supplied a
// connector_id. It must not run on Read/Update where connector_id is computed
// from the API after the resource exists.
func enforceUserSuppliedConnectorIDVersion(ctx context.Context, client entitycore.MinVersionClient, plan tfModel) diag.Diagnostics {
	if !userSuppliedConnectorID(plan) {
		return nil
	}

	var diags diag.Diagnostics
	supported, vDiags := client.EnforceMinVersion(ctx, MinVersionSupportingPreconfiguredIDs)
	diags.Append(vDiags...)
	if diags.HasError() {
		return diags
	}
	if !supported {
		diags.AddError("Unsupported server version", preconfiguredConnectorIDUnsupportedMessage())
	}
	return diags
}
