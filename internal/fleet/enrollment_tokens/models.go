package enrollment_tokens

import (
	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type enrollmentTokensModel struct {
	ID       types.String           `tfsdk:"id"`
	PolicyID types.String           `tfsdk:"policy_id"`
	Tokens   []enrollmentTokenModel `tfsdk:"tokens"`
}

type enrollmentTokenModel struct {
	KeyID     types.String `tfsdk:"key_id"`
	ApiKey    types.String `tfsdk:"api_key"`
	ApiKeyID  types.String `tfsdk:"api_key_id"`
	CreatedAt types.String `tfsdk:"created_at"`
	Name      types.String `tfsdk:"name"`
	Active    types.Bool   `tfsdk:"active"`
	PolicyID  types.String `tfsdk:"policy_id"`
}

func (model *enrollmentTokensModel) populateFromAPI(data []fleetapi.EnrollmentApiKey) {
	model.Tokens = make([]enrollmentTokenModel, 0, len(data))
	for _, token := range data {
		itemModel := enrollmentTokenModel{}
		itemModel.populateFromAPI(token)
		model.Tokens = append(model.Tokens, itemModel)
	}
}

func (model *enrollmentTokenModel) populateFromAPI(data fleetapi.EnrollmentApiKey) {
	model.KeyID = types.StringValue(data.Id)
	model.Active = types.BoolValue(data.Active)
	model.ApiKey = types.StringValue(data.ApiKey)
	model.ApiKeyID = types.StringValue(data.ApiKeyId)
	model.CreatedAt = types.StringValue(data.CreatedAt)
	model.Name = types.StringPointerValue(data.Name)
	model.PolicyID = types.StringPointerValue(data.PolicyId)
}
