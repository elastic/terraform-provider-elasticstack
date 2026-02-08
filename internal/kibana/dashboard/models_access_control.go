package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AccessControlValue maps to the access_control block
type AccessControlValue struct {
	AccessMode types.String `tfsdk:"access_mode"`
	Owner      types.String `tfsdk:"owner"`
}

type accessControlAPIPostModel = struct {
	AccessMode *kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode `json:"access_mode,omitempty"`
	Owner      *string                                                  `json:"owner,omitempty"`
}

type accessControlAPIPutModel = struct {
	AccessMode *kbapi.PutDashboardsIdJSONBodyDataAccessControlAccessMode `json:"access_mode,omitempty"`
	Owner      *string                                                   `json:"owner,omitempty"`
}

// ToCreateAPI converts the Terraform model to the POST API model
func (m *AccessControlValue) ToCreateAPI() *accessControlAPIPostModel {
	if m == nil {
		return nil
	}

	apiModel := &accessControlAPIPostModel{}

	if utils.IsKnown(m.AccessMode) {
		apiModel.AccessMode = utils.Pointer(kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode(m.AccessMode.ValueString()))
	}

	if utils.IsKnown(m.Owner) {
		apiModel.Owner = utils.Pointer(m.Owner.ValueString())
	}

	return apiModel
}

// ToUpdateAPI converts the Terraform model to the PUT API model
func (m *AccessControlValue) ToUpdateAPI() *accessControlAPIPutModel {
	createModel := m.ToCreateAPI()
	if createModel == nil {
		return nil
	}

	return &accessControlAPIPutModel{
		AccessMode: (*kbapi.PutDashboardsIdJSONBodyDataAccessControlAccessMode)(createModel.AccessMode),
		Owner:      createModel.Owner,
	}
}

// newAccessControlFromAPI maps the API response to the Terraform model
func newAccessControlFromAPI(accessMode *string, owner *string) *AccessControlValue {
	if accessMode == nil && owner == nil {
		return nil
	}

	return &AccessControlValue{
		AccessMode: types.StringPointerValue(accessMode),
		Owner:      types.StringPointerValue(owner),
	}
}
