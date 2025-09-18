package output

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	_ resource.Resource                 = &outputResource{}
	_ resource.ResourceWithConfigure    = &outputResource{}
	_ resource.ResourceWithImportState  = &outputResource{}
	_ resource.ResourceWithUpgradeState = &outputResource{}
)

var MinVersionOutputKafka = version.Must(version.NewVersion("8.13.0"))

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &outputResource{}
}

type outputResource struct {
	client *clients.ApiClient
}

func (r *outputResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *outputResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "fleet_output")
}

func (r *outputResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("output_id"), req, resp)
}

func (r *outputResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			// Legacy provider versions used a block for the `ssl` attribute which means it was stored as a list.
			// This upgrader migrates the list into a single object if available within the raw state
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				if req.RawState == nil || req.RawState.JSON == nil {
					resp.Diagnostics.AddError("Invalid raw state", "Raw state or JSON is nil")
					return
				}

				// Default to returning the original state if no changes are needed
				resp.DynamicValue = &tfprotov6.DynamicValue{
					JSON: req.RawState.JSON,
				}

				var stateMap map[string]interface{}
				err := json.Unmarshal(req.RawState.JSON, &stateMap)
				if err != nil {
					resp.Diagnostics.AddError("Failed to unmarshal raw state", err.Error())
					return
				}

				sslInterface, ok := stateMap["ssl"]
				if !ok {
					return
				}

				sslList, ok := sslInterface.([]any)
				if !ok {
					resp.Diagnostics.AddAttributeError(path.Root("ssl"),
						"Unexpected type for legacy ssl attribute",
						fmt.Sprintf("Expected []any, got %T", sslInterface),
					)
					return
				}

				if len(sslList) > 0 {
					stateMap["ssl"] = sslList[0]
				} else {
					delete(stateMap, "ssl")
				}

				stateJSON, err := json.Marshal(stateMap)
				if err != nil {
					resp.Diagnostics.AddError("Failed to marshal raw state", err.Error())
					return
				}

				resp.DynamicValue.JSON = stateJSON
			},
		},
	}
}
