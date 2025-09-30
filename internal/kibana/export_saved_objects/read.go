package export_saved_objects

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
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

	objectsList := utils.ListTypeToSlice(ctx, config.Objects, path.Root("objects"), &resp.Diagnostics, func(item objectModel, meta utils.ListMeta) struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	} {
		return struct {
			Id   string `json:"id"`
			Type string `json:"type"`
		}{
			Id:   item.ID.ValueString(),
			Type: item.Type.ValueString(),
		}
	})

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
