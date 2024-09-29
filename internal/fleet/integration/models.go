package integration

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Version     types.String `tfsdk:"version"`
	Force       types.Bool   `tfsdk:"force"`
	SkipDestroy types.Bool   `tfsdk:"skip_destroy"`
}

func getPackageID(name string, version string) string {
	hash, _ := utils.StringToHash(name + version)
	return *hash
}
