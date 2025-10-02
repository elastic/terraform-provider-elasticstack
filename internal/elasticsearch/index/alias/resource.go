package alias

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &aliasResource{}
var _ resource.ResourceWithConfigure = &aliasResource{}
var _ resource.ResourceWithImportState = &aliasResource{}
var _ resource.ResourceWithValidateConfig = &aliasResource{}

func NewAliasResource() resource.Resource {
	return &aliasResource{}
}

type aliasResource struct {
	client *clients.ApiClient
}

func (r *aliasResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_elasticsearch_alias"
}

func (r *aliasResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *aliasResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *aliasResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config tfModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that write_index doesn't appear in read_indices
	if config.WriteIndex.IsNull() {
		return
	}

	if config.ReadIndices.IsNull() {
		return
	}

	// Get the write index name
	var writeIndex indexModel
	diags := config.WriteIndex.As(ctx, &writeIndex, basetypes.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	writeIndexName := writeIndex.Name.ValueString()

	// Only validate if write index name is not empty
	if writeIndexName == "" {
		return
	}

	// Get all read indices
	var readIndices []indexModel
	if diags := config.ReadIndices.ElementsAs(ctx, &readIndices, false); !diags.HasError() {
		for _, readIndex := range readIndices {
			readIndexName := readIndex.Name.ValueString()
			if readIndexName != "" && readIndexName == writeIndexName {
				resp.Diagnostics.AddError(
					"Invalid Configuration",
					fmt.Sprintf("Index '%s' cannot be both a write index and a read index", writeIndexName),
				)
				return
			}
		}
	}
}
