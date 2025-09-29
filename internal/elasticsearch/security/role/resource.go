package role

import (
	"context"
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
			PriorSchema:   utils.Pointer(GetSchema(0)),
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

	if priorState["global"] == "" {
		delete(priorState, "global")
	}

	if priorState["metadata"] == "" {
		delete(priorState, "metadata")
	}

	indices, ok := priorState["indices"]
	if ok {
		indicesSlice, ok := indices.([]interface{})
		if ok {
			for i, index := range indicesSlice {
				indexMap, ok := index.(map[string]interface{})
				if ok {
					if indexMap["query"] == "" {
						delete(indexMap, "query")
					}
					indicesSlice[i] = indexMap
				}
			}
		}
	}

	remoteIndices, ok := priorState["remote_indices"]
	if ok {
		remoteIndicesSlice, ok := remoteIndices.([]interface{})
		if ok {
			for i, remoteIndex := range remoteIndicesSlice {
				remoteIndexMap, ok := remoteIndex.(map[string]interface{})
				if ok {
					if remoteIndexMap["query"] == "" {
						delete(remoteIndexMap, "query")
					}
					remoteIndicesSlice[i] = remoteIndexMap
				}
			}
		}
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
