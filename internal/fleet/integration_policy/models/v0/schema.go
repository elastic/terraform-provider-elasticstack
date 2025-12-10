package v0

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func GetSchema() *schema.Schema {
	return &schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"id":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"policy_id":           schema.StringAttribute{Computed: true, Optional: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()}},
			"name":                schema.StringAttribute{Required: true},
			"namespace":           schema.StringAttribute{Required: true},
			"agent_policy_id":     schema.StringAttribute{Required: true},
			"description":         schema.StringAttribute{Optional: true},
			"enabled":             schema.BoolAttribute{Computed: true, Optional: true, Default: booldefault.StaticBool(true)},
			"force":               schema.BoolAttribute{Optional: true},
			"integration_name":    schema.StringAttribute{Required: true},
			"integration_version": schema.StringAttribute{Required: true},
			"vars_json":           schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
		},
		Blocks: map[string]schema.Block{
			"input": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"input_id":     schema.StringAttribute{Required: true},
						"enabled":      schema.BoolAttribute{Computed: true, Optional: true, Default: booldefault.StaticBool(true)},
						"streams_json": schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
						"vars_json":    schema.StringAttribute{Computed: true, Optional: true, Sensitive: true},
					},
				},
			},
		},
	}
}
