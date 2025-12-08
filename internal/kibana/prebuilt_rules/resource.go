package prebuilt_rules

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource               = &PrebuiltRuleResource{}
	_ resource.ResourceWithConfigure  = &PrebuiltRuleResource{}
	_ resource.ResourceWithModifyPlan = &PrebuiltRuleResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &PrebuiltRuleResource{}
}

type PrebuiltRuleResource struct {
	client *clients.ApiClient
}

func (r *PrebuiltRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clients.ConvertProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *PrebuiltRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, "kibana_install_prebuilt_rules")
}

func (r *PrebuiltRuleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var model prebuiltRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !utils.IsKnown(model.ID) {
		// Resource is being created, nothing to modify
		return
	}

	updated := false
	if notInstalled := model.RulesNotInstalled.ValueInt64(); notInstalled > 0 {
		updated = true
		model.RulesNotInstalled = types.Int64Value(0)
		model.RulesInstalled = types.Int64Value(model.RulesInstalled.ValueInt64() + notInstalled)
	}

	if notUpdated := model.RulesNotUpdated.ValueInt64(); notUpdated > 0 {
		updated = true
		model.RulesNotUpdated = types.Int64Value(0)
	}

	if TimelinesNotInstalled := model.TimelinesNotInstalled.ValueInt64(); TimelinesNotInstalled > 0 {
		updated = true
		model.TimelinesNotInstalled = types.Int64Value(0)
		model.TimelinesInstalled = types.Int64Value(model.TimelinesInstalled.ValueInt64() + TimelinesNotInstalled)
	}

	if timelinesNotUpdated := model.TimelinesNotUpdated.ValueInt64(); timelinesNotUpdated > 0 {
		updated = true
		model.TimelinesNotUpdated = types.Int64Value(0)
	}

	if updated {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &model)...)
	}
}
