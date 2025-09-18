package export_saved_objects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read refreshes the Terraform state with the latest data.
func (d *dataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dataSourceModel

	// Read configuration
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get Kibana client
	oapiClient, err := d.client.GetKibanaOapiClient()
	if err != nil {
		resp.Diagnostics.AddError("unable to get Kibana client", err.Error())
		return
	}

	// Set default space_id if not provided
	spaceId := "default"
	if !config.SpaceID.IsNull() && !config.SpaceID.IsUnknown() {
		spaceId = config.SpaceID.ValueString()
	}

	// Parse objects JSON
	var objectsList kbapi.PostSavedObjectsExportJSONBodyHasReference1
	objectsJSON := config.Objects.ValueString()

	var rawObjects []map[string]interface{}
	if err := json.Unmarshal([]byte(objectsJSON), &rawObjects); err != nil {
		resp.Diagnostics.AddError("Invalid objects JSON", fmt.Sprintf("Error parsing objects JSON: %v", err))
		return
	}

	for _, obj := range rawObjects {
		id, ok := obj["id"].(string)
		if !ok {
			resp.Diagnostics.AddError("Invalid object", "Object missing 'id' field")
			return
		}
		objType, ok := obj["type"].(string)
		if !ok {
			resp.Diagnostics.AddError("Invalid object", "Object missing 'type' field")
			return
		}
		objectsList = append(objectsList, struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		}{
			Id:   id,
			Type: objType,
		})
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
	apiResp, err := oapiClient.API.PostSavedObjectsExportWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("API call failed", fmt.Sprintf("Unable to export saved objects: %v", err))
		return
	}

	if apiResp.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError(
			"Unexpected API response",
			fmt.Sprintf("Unexpected status code from server: got HTTP %d, response: %s", apiResp.StatusCode(), string(apiResp.Body)),
		)
		return
	}

	// Create composite ID for state tracking
	compositeID := &clients.CompositeId{ClusterId: spaceId, ResourceId: "export"}

	// Set the state
	var state dataSourceModel
	state.ID = types.StringValue(compositeID.String())
	state.SpaceID = types.StringValue(spaceId)
	state.Objects = config.Objects
	state.ExcludeExportDetails = types.BoolValue(excludeExportDetails)
	state.IncludeReferencesDeep = types.BoolValue(includeReferencesDeep)
	state.ExportedObjects = types.StringValue(string(apiResp.Body))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
