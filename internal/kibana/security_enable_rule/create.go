package security_enable_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *EnableRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(r.upsert(ctx, req.Plan, &resp.State)...)
}
