package security_list_data_streams

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SecurityListDataStreamsModel represents the Terraform state/config model for the
// kibana_security_list_data_streams resource. This resource manages the creation of
// .lists and .items data streams required for security lists and exceptions.
type SecurityListDataStreamsModel struct {
	ID            types.String `tfsdk:"id"`
	SpaceID       types.String `tfsdk:"space_id"`
	ListIndex     types.Bool   `tfsdk:"list_index"`
	ListItemIndex types.Bool   `tfsdk:"list_item_index"`
}

// fromAPIResponse populates the model from API response data.
// This helper method ensures consistency in how API responses are mapped to Terraform state.
func (m *SecurityListDataStreamsModel) fromAPIResponse(spaceID string, listIndex, listItemIndex bool) {
	m.ID = types.StringValue(spaceID)
	m.SpaceID = types.StringValue(spaceID)
	m.ListIndex = types.BoolValue(listIndex)
	m.ListItemIndex = types.BoolValue(listItemIndex)
}
