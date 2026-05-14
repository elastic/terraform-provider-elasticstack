package contracttest

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func summarizeDiags(diags diag.Diagnostics) string {
	if diags == nil {
		return ""
	}
	var b strings.Builder
	for _, d := range diags {
		if d.Severity() == diag.SeverityError || d.Severity() == diag.SeverityWarning {
			b.WriteString(d.Severity().String())
			b.WriteString(": ")
			b.WriteString(d.Summary())
			if dt := d.Detail(); dt != "" {
				b.WriteString(" — ")
				b.WriteString(dt)
			}
			b.WriteString("\n")
		}
	}
	s := strings.TrimSuffix(b.String(), "\n")
	if s == "" {
		return "(no diagnostics text)"
	}
	return s
}
