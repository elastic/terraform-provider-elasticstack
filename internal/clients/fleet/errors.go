package fleet

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// fromErr recreates the sdkdiag.FromErr functionality.
func fromErr(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(err.Error(), ""),
	}
}

func reportUnknownError(statusCode int, body []byte) diag.Diagnostics {
	return diag.Diagnostics{
		diag.NewErrorDiagnostic(
			fmt.Sprintf("Unexpected status code from server: got HTTP %d", statusCode),
			string(body),
		),
	}
}
