package diagutil

import (
	"context"

	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var contextDeadlineExceededDiags = FrameworkDiagFromError(context.DeadlineExceeded)

func ContainsContextDeadlineExceeded(ctx context.Context, diags fwdiag.Diagnostics) bool {
	if len(contextDeadlineExceededDiags) == 0 {
		tflog.Error(ctx, "Expected context deadline exceeded diagnostics to contain at least one error")
		return false
	}

	return diags.Contains(contextDeadlineExceededDiags[0])
}
