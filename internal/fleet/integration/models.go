package integration

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Version                   types.String `tfsdk:"version"`
	Force                     types.Bool   `tfsdk:"force"`
	Prerelease                types.Bool   `tfsdk:"prerelease"`
	IgnoreMappingUpdateErrors types.Bool   `tfsdk:"ignore_mapping_update_errors"`
	SkipDataStreamRollover    types.Bool   `tfsdk:"skip_data_stream_rollover"`
	IgnoreConstraints         types.Bool   `tfsdk:"ignore_constraints"`
	SkipDestroy               types.Bool   `tfsdk:"skip_destroy"`
	SpaceIds                  types.Set    `tfsdk:"space_ids"` //> string
}

func getPackageID(name string, version string) string {
	hash, _ := utils.StringToHash(name + version)
	return *hash
}
