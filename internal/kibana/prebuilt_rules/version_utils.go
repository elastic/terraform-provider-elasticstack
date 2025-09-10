package prebuilt_rules

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func validatePrebuiltRulesServer(serverVersion *version.Version, serverFlavor string) diag.Diagnostics {
	var serverlessFlavor = "serverless"
	var prebuiltRulesMinSupportedVersion = version.Must(version.NewVersion("8.0.0"))
	var diags diag.Diagnostics

	if serverVersion.LessThan(prebuiltRulesMinSupportedVersion) && serverFlavor != serverlessFlavor {
		diags.AddError("Prebuilt rules API not supported", fmt.Sprintf(`The prebuilt rules feature requires a minimum Elasticsearch version of "%s" or a serverless Kibana instance.`, prebuiltRulesMinSupportedVersion))
		return diags
	}

	return nil
}