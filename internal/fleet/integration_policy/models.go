package integration_policy

import (
	"context"
	"sort"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationPolicyModel struct {
	ID                 types.String         `tfsdk:"id"`
	PolicyID           types.String         `tfsdk:"policy_id"`
	Name               types.String         `tfsdk:"name"`
	Namespace          types.String         `tfsdk:"namespace"`
	AgentPolicyID      types.String         `tfsdk:"agent_policy_id"`
	Description        types.String         `tfsdk:"description"`
	Enabled            types.Bool           `tfsdk:"enabled"`
	Force              types.Bool           `tfsdk:"force"`
	IntegrationName    types.String         `tfsdk:"integration_name"`
	IntegrationVersion types.String         `tfsdk:"integration_version"`
	Input              types.List           `tfsdk:"input"` //> integrationPolicyInputModel
	VarsJson           jsontypes.Normalized `tfsdk:"vars_json"`
}

type integrationPolicyInputModel struct {
	InputID     types.String         `tfsdk:"input_id"`
	Enabled     types.Bool           `tfsdk:"enabled"`
	StreamsJson jsontypes.Normalized `tfsdk:"streams_json"`
	VarsJson    jsontypes.Normalized `tfsdk:"vars_json"`
}

func (model *integrationPolicyModel) populateFromAPI(ctx context.Context, data *fleetapi.PackagePolicy) diag.Diagnostics {
	if data == nil {
		return nil
	}

	var diags diag.Diagnostics

	model.ID = types.StringValue(data.Id)
	model.PolicyID = types.StringValue(data.Id)
	model.Name = types.StringValue(data.Name)
	model.Namespace = types.StringPointerValue(data.Namespace)
	model.AgentPolicyID = types.StringPointerValue(data.PolicyId)
	model.Description = types.StringPointerValue(data.Description)
	model.Enabled = types.BoolValue(data.Enabled)
	model.IntegrationName = types.StringValue(data.Package.Name)
	model.IntegrationVersion = types.StringValue(data.Package.Version)
	model.VarsJson = utils.MapToNormalizedType(utils.Deref(data.Vars), path.Root("vars_json"), &diags)

	model.populateInputFromAPI(ctx, data.Inputs, &diags)

	return diags
}

func (model *integrationPolicyModel) populateInputFromAPI(ctx context.Context, inputs map[string]fleetapi.PackagePolicyInput, diags *diag.Diagnostics) {
	newInputs := utils.TransformMapToSlice(inputs, path.Root("input"), diags,
		func(inputData fleetapi.PackagePolicyInput, meta utils.MapMeta) integrationPolicyInputModel {
			return integrationPolicyInputModel{
				InputID:     types.StringValue(meta.Key),
				Enabled:     types.BoolPointerValue(inputData.Enabled),
				StreamsJson: utils.MapToNormalizedType(utils.Deref(inputData.Streams), meta.Path.AtName("streams_json"), diags),
				VarsJson:    utils.MapToNormalizedType(utils.Deref(inputData.Vars), meta.Path.AtName("vars_json"), diags),
			}
		})
	if newInputs == nil {
		model.Input = types.ListNull(getInputTypeV1())
	} else {
		oldInputs := utils.ListTypeAs[integrationPolicyInputModel](ctx, model.Input, path.Root("input"), diags)

		sortInputs(newInputs, oldInputs)

		inputList, d := types.ListValueFrom(ctx, getInputTypeV1(), newInputs)
		diags.Append(d...)

		model.Input = inputList
	}
}

func (model integrationPolicyModel) toAPIModel(ctx context.Context, isUpdate bool) (fleetapi.PackagePolicyRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := fleetapi.PackagePolicyRequest{
		Description: model.Description.ValueStringPointer(),
		Force:       model.Force.ValueBoolPointer(),
		Name:        model.Name.ValueString(),
		Namespace:   model.Namespace.ValueStringPointer(),
		Package: fleetapi.PackagePolicyRequestPackage{
			Name:    model.IntegrationName.ValueString(),
			Version: model.IntegrationVersion.ValueString(),
		},
		PolicyId: model.AgentPolicyID.ValueStringPointer(),
		Vars:     utils.MapRef(utils.NormalizedTypeToMap[any](model.VarsJson, path.Root("vars_json"), &diags)),
	}

	if isUpdate {
		body.Id = model.ID.ValueStringPointer()
	}

	body.Inputs = utils.MapRef(utils.ListTypeToMap(ctx, model.Input, path.Root("input"), &diags,
		func(inputModel integrationPolicyInputModel, meta utils.ListMeta) (string, fleetapi.PackagePolicyRequestInput) {
			return inputModel.InputID.ValueString(), fleetapi.PackagePolicyRequestInput{
				Enabled: inputModel.Enabled.ValueBoolPointer(),
				Streams: utils.MapRef(utils.NormalizedTypeToMap[fleetapi.PackagePolicyRequestInputStream](inputModel.StreamsJson, meta.Path.AtName("streams_json"), &diags)),
				Vars:    utils.MapRef(utils.NormalizedTypeToMap[any](inputModel.VarsJson, meta.Path.AtName("vars_json"), &diags)),
			}
		}))

	return body, diags
}

// sortInputs will sort the 'incoming' list of input definitions based on
// the order of inputs defined in the 'existing' list. Inputs not present in
// 'existing' will be placed at the end of the list. Inputs are identified by
// their ID ('input_id'). The 'incoming' slice will be sorted in-place.
func sortInputs(incoming []integrationPolicyInputModel, existing []integrationPolicyInputModel) {
	if len(existing) == 0 {
		sort.Slice(incoming, func(i, j int) bool {
			return incoming[i].InputID.ValueString() < incoming[j].InputID.ValueString()
		})
		return
	}

	idToIndex := make(map[string]int, len(existing))
	for index, inputData := range existing {
		inputID := inputData.InputID.ValueString()
		idToIndex[inputID] = index
	}

	sort.Slice(incoming, func(i, j int) bool {
		iID := incoming[i].InputID.ValueString()
		iIdx, ok := idToIndex[iID]
		if !ok {
			return false
		}

		jID := incoming[j].InputID.ValueString()
		jIdx, ok := idToIndex[jID]
		if !ok {
			return true
		}

		return iIdx < jIdx
	})
}
