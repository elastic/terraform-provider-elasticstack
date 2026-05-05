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

package role

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	MinSupportedRemoteIndicesVersion = version.Must(version.NewVersion("8.10.0"))
	MinSupportedDescriptionVersion   = version.Must(version.NewVersion("8.15.0"))
)

func writeRole(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, data Data) (Data, diag.Diagnostics) {
	var diags diag.Diagnostics
	roleID := resourceID

	id, sdkDiags := client.ID(ctx, roleID)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	serverVersion, sdkDiags := client.ServerVersion(ctx)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	// Check version requirements
	if typeutils.IsKnown(data.Description) {
		if serverVersion.LessThan(MinSupportedDescriptionVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'description' is supported only for Elasticsearch v%s and above", MinSupportedDescriptionVersion.String()))
			var zero Data
			return zero, diags
		}
	}

	if typeutils.IsKnown(data.RemoteIndices) {
		var remoteIndicesList []RemoteIndexPermsData
		diags.Append(data.RemoteIndices.ElementsAs(ctx, &remoteIndicesList, false)...)
		if len(remoteIndicesList) > 0 && serverVersion.LessThan(MinSupportedRemoteIndicesVersion) {
			diags.AddError("Unsupported Feature", fmt.Sprintf("'remote_indices' is supported only for Elasticsearch v%s and above", MinSupportedRemoteIndicesVersion.String()))
			var zero Data
			return zero, diags
		}
	}

	// Convert to API model
	role, modelDiags := data.toAPIModel(ctx)
	diags.Append(modelDiags...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	// Put the role
	sdkDiags = elasticsearch.PutRole(ctx, client, roleID, role)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	data.ID = types.StringValue(id.String())
	readData, found, readDiags := readRole(ctx, client, roleID, data)
	diags.Append(readDiags...)
	if diags.HasError() {
		var zero Data
		return zero, diags
	}

	if !found {
		diags.AddError("Not Found", fmt.Sprintf("Role %q was not found after update", roleID))
		var zero Data
		return zero, diags
	}

	return readData, diags
}
