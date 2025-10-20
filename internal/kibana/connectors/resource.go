package connectors

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithConfigure = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}

var (
	MinVersionSupportingPreconfiguredIDs = version.Must(version.NewVersion("8.8.0"))
)

type Resource struct {
	client *clients.ApiClient
}

func (r *Resource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(request.ProviderData)
	response.Diagnostics.Append(diags...)
	r.client = client
}

func (r *Resource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_kibana_action_connector"
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
}

func (r *Resource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {StateUpgrader: upgradeV0},
	}
}

// The schema between V0 and V1 is mostly the same, however config saved ""
// values to the state when null values were in the config. jsontypes.Normalized
// correctly states this is invalid JSON.
func upgradeV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var state map[string]interface{}

	removeEmptyString := func(state map[string]interface{}, key string) map[string]interface{} {
		value, ok := state[key]
		if !ok {
			return state
		}

		valueString, ok := value.(string)
		if !ok || valueString != "" {
			return state
		}

		delete(state, key)
		return state
	}

	err := json.Unmarshal(req.RawState.JSON, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal state", err.Error())
		return
	}

	state = removeEmptyString(state, "config")
	state = removeEmptyString(state, "secrets")

	stateBytes, err := json.Marshal(state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal state", err.Error())
		return
	}

	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateBytes,
	}
}
