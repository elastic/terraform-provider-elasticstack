package api_key

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (r Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planModel tfModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, diags := clients.MaybeNewApiClientFromFrameworkResource(ctx, planModel.ElasticsearchConnection, r.client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel, diags := r.buildApiModel(ctx, planModel, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	putResponse, diags := elasticsearch.CreateApiKey(client, &apiModel)
	resp.Diagnostics.Append(diags...)
	if putResponse == nil || resp.Diagnostics.HasError() {
		return
	}

	id, sdkDiags := client.ID(ctx, putResponse.Id)
	resp.Diagnostics.Append(utils.FrameworkDiagsFromSDK(sdkDiags)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCreate(*putResponse)
	resp.Diagnostics.Append(resp.State.Set(ctx, planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	finalModel, diags := r.read(ctx, client, planModel)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, *finalModel)...)
}

func (r Resource) buildApiModel(ctx context.Context, model tfModel, client *clients.ApiClient) (models.ApiKey, diag.Diagnostics) {
	apiModel, diags := model.toAPIModel()
	if diags.HasError() {
		return models.ApiKey{}, diags
	}

	hasRestriction := false
	keysWithRestrictions := []string{}
	for key, descriptor := range apiModel.RolesDescriptors {
		if descriptor.Restriction != nil {
			hasRestriction = true
			keysWithRestrictions = append(keysWithRestrictions, key)
		}
	}

	if hasRestriction {
		isSupported, diags := doesCurrentVersionSupportRestrictionOnApiKey(ctx, client)
		if diags.HasError() {
			return models.ApiKey{}, diags
		}

		if !isSupported {
			diags.AddAttributeError(
				path.Root("roles_descriptors"),
				"Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch",
				fmt.Sprintf("Specifying `restriction` on an API key role description is not supported in this version of Elasticsearch. Role descriptor(s) %s", strings.Join(keysWithRestrictions, ", ")),
			)
			return models.ApiKey{}, diags
		}
	}

	return apiModel, nil
}

func doesCurrentVersionSupportRestrictionOnApiKey(ctx context.Context, client *clients.ApiClient) (bool, diag.Diagnostics) {
	currentVersion, diags := client.ServerVersion(ctx)

	if diags.HasError() {
		return false, utils.FrameworkDiagsFromSDK(diags)
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithRestriction), nil
}
