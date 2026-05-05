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

package exportsavedobjects

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// readDataSource is the envelope read callback for the export saved objects data source.
func readDataSource(ctx context.Context, kbClient *clients.KibanaScopedClient, config dataSourceModel) (dataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Get Kibana client
	oapiClient, err := kbClient.GetKibanaOapiClient()
	if err != nil {
		diags.AddError("unable to get Kibana client", err.Error())
		return config, diags
	}

	// Set default space_id if not provided
	spaceID := "default"
	if !config.SpaceID.IsNull() && !config.SpaceID.IsUnknown() {
		spaceID = config.SpaceID.ValueString()
	}

	objectsList := typeutils.ListTypeToSlice(ctx, config.Objects, path.Root("objects"), &diags, func(item objectModel, _ typeutils.ListMeta) struct {
		//nolint:revive
		Id   string `json:"id"`
		Type string `json:"type"`
	} {
		return struct {
			//nolint:revive
			Id   string `json:"id"`
			Type string `json:"type"`
		}{
			Id:   item.ID.ValueString(),
			Type: item.Type.ValueString(),
		}
	})
	if diags.HasError() {
		return config, diags
	}

	// Set default values for boolean options
	excludeExportDetails := true
	if !config.ExcludeExportDetails.IsNull() && !config.ExcludeExportDetails.IsUnknown() {
		excludeExportDetails = config.ExcludeExportDetails.ValueBool()
	}

	includeReferencesDeep := true
	if !config.IncludeReferencesDeep.IsNull() && !config.IncludeReferencesDeep.IsUnknown() {
		includeReferencesDeep = config.IncludeReferencesDeep.ValueBool()
	}

	// Create request body
	body := kbapi.PostSavedObjectsExportJSONRequestBody{
		ExcludeExportDetails:  &excludeExportDetails,
		IncludeReferencesDeep: &includeReferencesDeep,
		Objects:               &objectsList,
	}

	// Make the API call
	apiResp, err := oapiClient.API.PostSavedObjectsExportWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		diags.AddError("API call failed", fmt.Sprintf("Unable to export saved objects: %v", err))
		return config, diags
	}

	if apiResp.StatusCode() != http.StatusOK {
		diags.AddError(
			"Unexpected API response",
			fmt.Sprintf("Unexpected status code from server: got HTTP %d, response: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return config, diags
	}

	// Create composite ID for state tracking
	compositeID := &clients.CompositeID{ClusterID: spaceID, ResourceID: "export"}

	config.ID = types.StringValue(compositeID.String())
	config.SpaceID = types.StringValue(spaceID)
	config.ExcludeExportDetails = types.BoolValue(excludeExportDetails)
	config.IncludeReferencesDeep = types.BoolValue(includeReferencesDeep)
	config.ExportedObjects = types.StringValue(string(apiResp.Body))

	return config, diags
}
