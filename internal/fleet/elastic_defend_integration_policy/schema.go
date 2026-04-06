// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package elasticdefendintegrationpolicy

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func resourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Elastic Defend Fleet integration policy (package policy for the `endpoint` package). " +
			"Uses a two-phase create (bootstrap then finalize) and preserves server-managed payloads such as " +
			"`artifact_manifest` and the package policy `version` in private state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Unique identifier of the Elastic Defend integration policy. Used as the import key.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the integration policy.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "The namespace of the integration policy.",
				Required:    true,
			},
			"agent_policy_id": schema.StringAttribute{
				Description: "ID of the agent policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the integration policy.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Enable the integration policy.",
				Computed:    true,
				Optional:    true,
				Default:     booldefault.StaticBool(true),
			},
			"force": schema.BoolAttribute{
				Description: "Force operations, such as creation and deletion, to occur.",
				Optional:    true,
			},
			"integration_version": schema.StringAttribute{
				Description: "The version of the Elastic Defend integration package.",
				Required:    true,
			},
			"space_ids": schema.SetAttribute{
				Description: "The Kibana space IDs where this integration policy is available. " +
					"When set, must match the space_ids of the referenced agent policy. " +
					"If not set, will be inherited from the agent policy.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"preset": schema.StringAttribute{
				Description: "Elastic Defend preset configuration. Maps to `endpointConfig.preset` in the Defend API. " +
					"Common values include `\"NGAv1\"`, `\"NGAV\"`, `\"dataCollection\"`, `\"EDRComplete\"`, `\"EDREssential\"`.",
				Optional: true,
			},
			"policy": policySchema(),
		},
	}
}

func policySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Elastic Defend policy configuration.",
		Required:    true,
		Attributes: map[string]schema.Attribute{
			"windows": windowsPolicySchema(),
			"mac":     macPolicySchema(),
			"linux":   linuxPolicySchema(),
		},
	}
}

func popupItemSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Optional: true,
		Default:  objectdefault.StaticValue(popupItemDefaultValue()),
		Attributes: map[string]schema.Attribute{
			"message": schema.StringAttribute{
				Description: "The popup message text.",
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the popup notification is enabled.",
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Optional:    true,
			},
		},
	}
}

func protectionModeSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:    true,
		Optional:    true,
		Description: description,
		Default:     objectdefault.StaticValue(protectionModeDefaultValue()),
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Description: "Protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`.",
				Computed:    true,
				Default:     stringdefault.StaticString("off"),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("off", "detect", "prevent"),
				},
			},
			"supported": schema.BoolAttribute{
				Description: "Whether this protection is supported on the platform.",
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Optional:    true,
			},
		},
	}
}

func behaviorProtectionSchema(description string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed:    true,
		Optional:    true,
		Description: description,
		Default:     objectdefault.StaticValue(behaviorProtectionDefaultValue()),
		Attributes: map[string]schema.Attribute{
			"mode": schema.StringAttribute{
				Description: "Protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`.",
				Computed:    true,
				Default:     stringdefault.StaticString("off"),
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("off", "detect", "prevent"),
				},
			},
			"supported": schema.BoolAttribute{
				Description: "Whether this protection is supported on the platform.",
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Optional:    true,
			},
			"reputation_service": schema.BoolAttribute{
				Description: "Whether reputation service is enabled.",
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Optional:    true,
			},
		},
	}
}

func popupItemDefaultValue() types.Object {
	return types.ObjectValueMust(popupItemAttrTypes(), map[string]attr.Value{
		"message": types.StringValue(""),
		"enabled": types.BoolValue(false),
	})
}

func protectionModeDefaultValue() types.Object {
	return types.ObjectValueMust(protectionModeAttrTypes(), map[string]attr.Value{
		"mode":      types.StringValue("off"),
		"supported": types.BoolValue(true),
	})
}

func behaviorProtectionDefaultValue() types.Object {
	return types.ObjectValueMust(behaviorProtectionAttrTypes(), map[string]attr.Value{
		"mode":               types.StringValue("off"),
		"supported":          types.BoolValue(true),
		"reputation_service": types.BoolValue(false),
	})
}

func antivirusRegistrationDefaultValue() types.Object {
	return types.ObjectValueMust(antivirusRegistrationAttrTypes(), map[string]attr.Value{
		"mode":    types.StringValue("disabled"),
		"enabled": types.BoolValue(false),
	})
}

func credentialHardeningDefaultValue() types.Object {
	return types.ObjectValueMust(credentialHardeningAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(false),
	})
}

func attackSurfaceReductionDefaultValue() types.Object {
	return types.ObjectValueMust(attackSurfaceReductionAttrTypes(), map[string]attr.Value{
		"credential_hardening": credentialHardeningDefaultValue(),
	})
}

func windowsPopupDefaultValue() types.Object {
	return types.ObjectValueMust(windowsPopupAttrTypes(), map[string]attr.Value{
		"malware":             popupItemDefaultValue(),
		"ransomware":          popupItemDefaultValue(),
		"memory_protection":   popupItemDefaultValue(),
		"behavior_protection": popupItemDefaultValue(),
	})
}

func windowsPolicySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Windows-specific Elastic Defend policy settings.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"events": schema.SingleNestedAttribute{
				Description: "Windows event collection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"process": schema.BoolAttribute{
						Description: "Collect process events.",
						Optional:    true,
					},
					"network": schema.BoolAttribute{
						Description: "Collect network events.",
						Optional:    true,
					},
					"file": schema.BoolAttribute{
						Description: "Collect file events.",
						Optional:    true,
					},
					"dll_and_driver_load": schema.BoolAttribute{
						Description: "Collect DLL and driver load events.",
						Optional:    true,
					},
					"dns": schema.BoolAttribute{
						Description: "Collect DNS events.",
						Optional:    true,
					},
					"registry": schema.BoolAttribute{
						Description: "Collect registry events.",
						Optional:    true,
					},
					"security": schema.BoolAttribute{
						Description: "Collect security events.",
						Optional:    true,
					},
					"authentication": schema.BoolAttribute{
						Description: "Collect authentication events.",
						Optional:    true,
					},
				},
			},
			"malware": schema.SingleNestedAttribute{
				Description: "Windows malware protection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Malware protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("off", "detect", "prevent"),
						},
					},
					"blocklist": schema.BoolAttribute{
						Description: "Whether blocklist is enabled.",
						Optional:    true,
					},
					"on_write_scan": schema.BoolAttribute{
						Description: "Whether on-write scan is enabled.",
						Optional:    true,
					},
					"notify_user": schema.BoolAttribute{
						Description: "Whether to notify the user on malware detection.",
						Optional:    true,
					},
				},
			},
			"ransomware":          protectionModeSchema("Windows ransomware protection settings."),
			"memory_protection":   protectionModeSchema("Windows memory protection settings."),
			"behavior_protection": behaviorProtectionSchema("Windows behavior protection settings."),
			"popup": schema.SingleNestedAttribute{
				Description: "Windows popup notification settings.",
				Computed:    true,
				Optional:    true,
				Default:     objectdefault.StaticValue(windowsPopupDefaultValue()),
				Attributes: map[string]schema.Attribute{
					"malware":             popupItemSchema(),
					"ransomware":          popupItemSchema(),
					"memory_protection":   popupItemSchema(),
					"behavior_protection": popupItemSchema(),
				},
			},
			"logging": schema.SingleNestedAttribute{
				Description: "Windows logging settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"file": schema.StringAttribute{
						Description: "Log level for file logging. Valid values: `\"info\"`, `\"debug\"`, `\"warning\"`, `\"error\"`, `\"critical\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("info", "debug", "warning", "error", "critical"),
						},
					},
				},
			},
			"antivirus_registration": schema.SingleNestedAttribute{
				Description: "Windows antivirus registration settings.",
				Computed:    true,
				Optional:    true,
				Default:     objectdefault.StaticValue(antivirusRegistrationDefaultValue()),
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Antivirus registration mode. Valid values: `\"enabled\"`, `\"disabled\"`, `\"sync_with_malware_prevent\"`.",
						Computed:    true,
						Default:     stringdefault.StaticString("disabled"),
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("enabled", "disabled", "sync_with_malware_prevent"),
						},
					},
					"enabled": schema.BoolAttribute{
						Description: "Whether antivirus registration is enabled.",
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Optional:    true,
					},
				},
			},
			"attack_surface_reduction": schema.SingleNestedAttribute{
				Description: "Windows attack surface reduction settings.",
				Computed:    true,
				Optional:    true,
				Default:     objectdefault.StaticValue(attackSurfaceReductionDefaultValue()),
				Attributes: map[string]schema.Attribute{
					"credential_hardening": schema.SingleNestedAttribute{
						Description: "Credential hardening settings.",
						Computed:    true,
						Optional:    true,
						Default:     objectdefault.StaticValue(credentialHardeningDefaultValue()),
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Description: "Whether credential hardening is enabled.",
								Computed:    true,
								Default:     booldefault.StaticBool(false),
								Optional:    true,
							},
						},
					},
				},
			},
		},
	}
}

func macPolicySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "macOS-specific Elastic Defend policy settings.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"events": schema.SingleNestedAttribute{
				Description: "macOS event collection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"process": schema.BoolAttribute{
						Description: "Collect process events.",
						Optional:    true,
					},
					"network": schema.BoolAttribute{
						Description: "Collect network events.",
						Optional:    true,
					},
					"file": schema.BoolAttribute{
						Description: "Collect file events.",
						Optional:    true,
					},
				},
			},
			"malware": schema.SingleNestedAttribute{
				Description: "macOS malware protection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Malware protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("off", "detect", "prevent"),
						},
					},
					"blocklist": schema.BoolAttribute{
						Description: "Whether blocklist is enabled.",
						Optional:    true,
					},
					"on_write_scan": schema.BoolAttribute{
						Description: "Whether on-write scan is enabled.",
						Optional:    true,
					},
					"notify_user": schema.BoolAttribute{
						Description: "Whether to notify the user on malware detection.",
						Optional:    true,
					},
				},
			},
			"memory_protection":   protectionModeSchema("macOS memory protection settings."),
			"behavior_protection": behaviorProtectionSchema("macOS behavior protection settings."),
			"popup": schema.SingleNestedAttribute{
				Description: "macOS popup notification settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"malware":             popupItemSchema(),
					"memory_protection":   popupItemSchema(),
					"behavior_protection": popupItemSchema(),
				},
			},
			"logging": schema.SingleNestedAttribute{
				Description: "macOS logging settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"file": schema.StringAttribute{
						Description: "Log level for file logging. Valid values: `\"info\"`, `\"debug\"`, `\"warning\"`, `\"error\"`, `\"critical\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("info", "debug", "warning", "error", "critical"),
						},
					},
				},
			},
		},
	}
}

func linuxPolicySchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Description: "Linux-specific Elastic Defend policy settings.",
		Optional:    true,
		Attributes: map[string]schema.Attribute{
			"events": schema.SingleNestedAttribute{
				Description: "Linux event collection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"process": schema.BoolAttribute{
						Description: "Collect process events.",
						Optional:    true,
					},
					"network": schema.BoolAttribute{
						Description: "Collect network events.",
						Optional:    true,
					},
					"file": schema.BoolAttribute{
						Description: "Collect file events.",
						Optional:    true,
					},
					"session_data": schema.BoolAttribute{
						Description: "Collect session data events.",
						Optional:    true,
					},
					"tty_io": schema.BoolAttribute{
						Description: "Collect TTY I/O events.",
						Optional:    true,
					},
				},
			},
			"malware": schema.SingleNestedAttribute{
				Description: "Linux malware protection settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"mode": schema.StringAttribute{
						Description: "Malware protection mode. Valid values: `\"off\"`, `\"detect\"`, `\"prevent\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("off", "detect", "prevent"),
						},
					},
					"blocklist": schema.BoolAttribute{
						Description: "Whether blocklist is enabled.",
						Optional:    true,
					},
				},
			},
			"memory_protection":   protectionModeSchema("Linux memory protection settings."),
			"behavior_protection": behaviorProtectionSchema("Linux behavior protection settings."),
			"popup": schema.SingleNestedAttribute{
				Description: "Linux popup notification settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"malware":             popupItemSchema(),
					"memory_protection":   popupItemSchema(),
					"behavior_protection": popupItemSchema(),
				},
			},
			"logging": schema.SingleNestedAttribute{
				Description: "Linux logging settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"file": schema.StringAttribute{
						Description: "Log level for file logging. Valid values: `\"info\"`, `\"debug\"`, `\"warning\"`, `\"error\"`, `\"critical\"`.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("info", "debug", "warning", "error", "critical"),
						},
					},
				},
			},
		},
	}
}
