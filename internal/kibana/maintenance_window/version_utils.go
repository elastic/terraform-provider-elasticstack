package maintenance_window

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func validateMaintenanceWindowServer(serverVersion *version.Version, serverFlavor string) diag.Diagnostics {
	var serverlessFlavor = "serverless"
	var maintenanceWindowPublicAPIMinSupportedVersion = version.Must(version.NewVersion("9.1.0"))
	var diags diag.Diagnostics

	if serverVersion.LessThan(maintenanceWindowPublicAPIMinSupportedVersion) && serverFlavor != serverlessFlavor {
		diags.AddError("Maintenance window API not supported", fmt.Sprintf(`The maintenance Window public API feature requires a minimum Elasticsearch version of "%s" or a serverless Kibana instance.`, maintenanceWindowPublicAPIMinSupportedVersion))
		return diags
	}

	return nil
}
