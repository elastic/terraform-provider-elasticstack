package role

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &roleResource{}
var _ resource.ResourceWithConfigure = &roleResource{}
var _ resource.ResourceWithImportState = &roleResource{}
var _ resource.ResourceWithUpgradeState = &roleResource{}

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *clients.ApiClient
}

func (r *roleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_security_role"
}

func (r *roleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *roleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: v0ToV1,
		},
	}
}

func v0ToV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var priorState map[string]interface{}
	err := json.Unmarshal(req.RawState.JSON, &priorState)
	if err != nil {
		resp.Diagnostics.AddError("State Upgrade Error", "Could not unmarshal prior state: "+err.Error())
		return
	}

	if global := priorState["global"]; global == nil || global == "" {
		delete(priorState, "global")
	}

	if metadata := priorState["metadata"]; metadata == nil || metadata == "" {
		delete(priorState, "metadata")
	}

	indices, ok := priorState["indices"]
	if ok {
		priorState["indices"] = convertV0Indices(indices)
	}

	remoteIndices, ok := priorState["remote_indices"]
	if ok {
		priorState["remote_indices"] = convertV0Indices(remoteIndices)
	}

	stateJSON, err := json.Marshal(priorState)
	if err != nil {
		resp.Diagnostics.AddError("State Upgrade Error", "Could not marshal new state: "+err.Error())
		return
	}
	resp.DynamicValue = &tfprotov6.DynamicValue{
		JSON: stateJSON,
	}
}

func convertV0Indices(indices interface{}) interface{} {
	indicesSlice, ok := indices.([]interface{})
	if ok {
		for i, index := range indicesSlice {
			indexMap, ok := index.(map[string]interface{})
			if ok {
				if indexMap["query"] == "" {
					delete(indexMap, "query")
				}
				// Convert field_security from a list to an object
				if fs, ok := indexMap["field_security"]; ok {
					fsList, ok := fs.([]interface{})
					if ok && len(fsList) > 0 {
						indexMap["field_security"] = fsList[0]
					} else {
						delete(indexMap, "field_security")
					}
				}
				indicesSlice[i] = indexMap
			}
		}
	}
	return indicesSlice
}
