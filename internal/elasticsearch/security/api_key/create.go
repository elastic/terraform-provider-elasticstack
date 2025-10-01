package api_key

import (
	"context"
	"fmt"
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
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

	if planModel.Type.ValueString() == "cross_cluster" {
		createDiags := r.createCrossClusterApiKey(ctx, client, &planModel)
		resp.Diagnostics.Append(createDiags...)
	} else {
		createDiags := r.createApiKey(ctx, client, &planModel)
		resp.Diagnostics.Append(createDiags...)
	}

	if resp.Diagnostics.HasError() {
		return
	}

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
		return false, diagutil.FrameworkDiagsFromSDK(diags)
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithRestriction), nil
}

func doesCurrentVersionSupportCrossClusterApiKey(ctx context.Context, client *clients.ApiClient) (bool, diag.Diagnostics) {
	currentVersion, diags := client.ServerVersion(ctx)

	if diags.HasError() {
		return false, diagutil.FrameworkDiagsFromSDK(diags)
	}

	return currentVersion.GreaterThanOrEqual(MinVersionWithCrossCluster), nil
}

func (r *Resource) createCrossClusterApiKey(ctx context.Context, client *clients.ApiClient, planModel *tfModel) diag.Diagnostics {
	// Check if the current version supports cross-cluster API keys
	isSupported, diags := doesCurrentVersionSupportCrossClusterApiKey(ctx, client)
	if diags.HasError() {
		return diags
	}
	if !isSupported {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Cross-cluster API keys not supported",
				fmt.Sprintf("Cross-cluster API keys are only supported in Elasticsearch version %s and above.", MinVersionWithCrossCluster.String()),
			),
		}
	}

	// Handle cross-cluster API key creation
	crossClusterModel, diags := planModel.toCrossClusterAPIModel(ctx)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateCrossClusterApiKey(client, &crossClusterModel)
	if createDiags.HasError() {
		return diag.Diagnostics(createDiags)
	}
	if putResponse == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("API Key Creation Failed", "Cross-cluster API key creation returned nil response"),
		}
	}

	id, sdkDiags := client.ID(ctx, putResponse.Id)
	if sdkDiags.HasError() {
		return diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCrossClusterCreate(*putResponse)
	return nil
}

func (r *Resource) createApiKey(ctx context.Context, client *clients.ApiClient, planModel *tfModel) diag.Diagnostics {
	// Handle regular API key creation
	apiModel, diags := r.buildApiModel(ctx, *planModel, client)
	if diags.HasError() {
		return diags
	}

	putResponse, createDiags := elasticsearch.CreateApiKey(client, &apiModel)
	if createDiags.HasError() {
		return diag.Diagnostics(createDiags)
	}
	if putResponse == nil {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic("API Key Creation Failed", "API key creation returned nil response"),
		}
	}

	id, sdkDiags := client.ID(ctx, putResponse.Id)
	if sdkDiags.HasError() {
		return diagutil.FrameworkDiagsFromSDK(sdkDiags)
	}

	planModel.ID = basetypes.NewStringValue(id.String())
	planModel.populateFromCreate(*putResponse)
	return nil
}
