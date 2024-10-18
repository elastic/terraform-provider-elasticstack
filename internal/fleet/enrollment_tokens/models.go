package enrollment_tokens

import (
	"context"

	fleetapi "github.com/elastic/terraform-provider-elasticstack/generated/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type enrollmentTokensModel struct {
	ID       types.String `tfsdk:"id"`
	PolicyID types.String `tfsdk:"policy_id"`
	Tokens   types.List   `tfsdk:"tokens"` //> enrollmentTokenModel
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

func (model *enrollmentTokensModel) populateFromAPI(ctx context.Context, data []fleetapi.EnrollmentApiKey) (diags diag.Diagnostics) {
	model.Tokens = utils.SliceToListType(ctx, data, getTokenType(), path.Root("tokens"), &diags, newEnrollmentTokenModel)
	return
}

func newEnrollmentTokenModel(data fleetapi.EnrollmentApiKey, meta utils.ListMeta) enrollmentTokenModel {
	return enrollmentTokenModel{
		KeyID:     types.StringValue(data.Id),
		Active:    types.BoolValue(data.Active),
		ApiKey:    types.StringValue(data.ApiKey),
		ApiKeyID:  types.StringValue(data.ApiKeyId),
		CreatedAt: types.StringValue(data.CreatedAt),
		Name:      types.StringPointerValue(data.Name),
		PolicyID:  types.StringPointerValue(data.PolicyId),
	}
}
