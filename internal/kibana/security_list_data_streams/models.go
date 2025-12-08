package security_list_data_streams

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityListDataStreamsModel struct {
	ID           types.String `tfsdk:"id"`
	SpaceID      types.String `tfsdk:"space_id"`
	Acknowledged types.Bool   `tfsdk:"acknowledged"`
}
