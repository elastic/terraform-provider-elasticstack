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

package policyshape

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// InputVarsOptions configures the one attribute that intentionally differs
// in shape between integration_policy and managedintegration: the
// input-level "vars" attribute. managedintegration marks it Computed with
// UseStateForUnknown because some packages (e.g. cloud_security_posture)
// populate informational input-level vars that are always present in the API
// response regardless of configuration; integration_policy's vars attribute
// is purely user-supplied (Optional only). This is preserved deliberately
// rather than unified away, so callers must say explicitly which behavior
// they want.
type InputVarsOptions struct {
	// Computed marks "vars" Computed and adds stringplanmodifier.UseStateForUnknown.
	Computed bool
	// Description is this resource's wording for the input-level "vars"
	// attribute (the two callers' text differs: managedintegration's
	// explains the Computed behavior above).
	Description string
}

// InputsNestedObjectConfig carries the resource-specific pieces that
// InputsNestedObject needs in order to serve both integration_policy (which
// includes a package-computed "defaults" sub-object and plain-text
// Description fields) and managedintegration (which omits "defaults" and
// uses MarkdownDescription fields throughout), without losing either
// resource's existing, intentional behavior.
type InputsNestedObjectConfig struct {
	// CustomType is the nested object's CustomType. Each caller wraps a
	// different attribute-types map in its own InputType (integration_policy
	// includes "defaults"; managedintegration doesn't), so this is supplied
	// by the caller rather than built here.
	CustomType basetypes.ObjectTypable
	// VarsAreSensitive marks every "vars" attribute (input- and
	// stream-level) Sensitive.
	VarsAreSensitive bool
	// UseMarkdownDescription selects MarkdownDescription over Description
	// for every field this builder sets, matching the caller's existing
	// per-resource convention (managedintegration uses Markdown throughout;
	// integration_policy uses plain Description throughout).
	UseMarkdownDescription bool
	// IncludeDefaults adds the package-computed "defaults" sub-object nested
	// under each input (integration_policy only).
	IncludeDefaults bool
	// InputVars configures the input-level "vars" attribute.
	InputVars InputVarsOptions
}

// splitDescription routes description text to either the plain Description
// or MarkdownDescription return slot, matching useMarkdown.
func splitDescription(text string, useMarkdown bool) (description, markdownDescription string) {
	if useMarkdown {
		return "", text
	}
	return text, ""
}

// InputsNestedObject builds the "inputs" map's NestedAttributeObject shared
// by integration_policy and managedintegration: enabled/condition/vars/
// streams, plus an optional package-computed "defaults" sub-object. See
// InputsNestedObjectConfig for the pieces that intentionally differ between
// those two callers.
func InputsNestedObject(cfg InputsNestedObjectConfig) schema.NestedAttributeObject {
	enabledDescription, enabledMarkdown := splitDescription("Enable the input.", cfg.UseMarkdownDescription)
	conditionDescription, conditionMarkdown := splitDescription(
		"Agent condition expression to evaluate whether to apply this input.", cfg.UseMarkdownDescription)
	streamsDescription, streamsMarkdown := splitDescription("Input streams mapped by stream ID.", cfg.UseMarkdownDescription)
	varsDescription, varsMarkdown := splitDescription(cfg.InputVars.Description, cfg.UseMarkdownDescription)

	varsAttribute := schema.StringAttribute{
		Description:         varsDescription,
		MarkdownDescription: varsMarkdown,
		CustomType:          jsontypes.NormalizedType{},
		Optional:            true,
		Sensitive:           cfg.VarsAreSensitive,
	}
	if cfg.InputVars.Computed {
		varsAttribute.Computed = true
		varsAttribute.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
	}

	attributes := map[string]schema.Attribute{
		AttrEnabled: schema.BoolAttribute{
			Description:         enabledDescription,
			MarkdownDescription: enabledMarkdown,
			Computed:            true,
			Optional:            true,
			Default:             booldefault.StaticBool(true),
		},
		AttrCondition: schema.StringAttribute{
			Description:         conditionDescription,
			MarkdownDescription: conditionMarkdown,
			Optional:            true,
		},
		AttrVars: varsAttribute,
		AttrStreams: schema.MapNestedAttribute{
			Description:         streamsDescription,
			MarkdownDescription: streamsMarkdown,
			Optional:            true,
			NestedObject: InputStreamNestedObject(
				cfg.VarsAreSensitive, cfg.UseMarkdownDescription),
		},
	}

	if cfg.IncludeDefaults {
		attributes[AttrDefaults] = inputDefaultsAttribute()
	}

	return schema.NestedAttributeObject{
		CustomType: cfg.CustomType,
		Attributes: attributes,
	}
}

// inputDefaultsAttribute builds the package-computed "defaults" sub-object
// nested under an input (integration_policy only; see
// InputDefaultsAttributeTypes for the matching attribute-types map).
func inputDefaultsAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Description: "Input defaults.",
		Computed:    true,
		Default: objectdefault.StaticValue(basetypes.NewObjectNull(
			InputDefaultsAttributeTypes(),
		)),
		Attributes: map[string]schema.Attribute{
			AttrVars: schema.StringAttribute{
				Description: "Input-level variable defaults as JSON.",
				CustomType:  jsontypes.NormalizedType{},
				Computed:    true,
			},
			AttrStreams: schema.MapNestedAttribute{
				Description: "Stream-level defaults mapped by stream ID.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						AttrEnabled: schema.BoolAttribute{
							Description: "Default enabled state for the stream.",
							Computed:    true,
						},
						AttrVars: schema.StringAttribute{
							Description: "Stream-level variable defaults as JSON.",
							CustomType:  jsontypes.NormalizedType{},
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// InputStreamNestedObject builds the "streams" map's NestedAttributeObject
// shared by integration_policy and managedintegration: enabled/condition/
// vars. Unlike the input-level vars attribute, stream-level vars has no
// Computed/UseStateForUnknown variant between the two callers, so it takes
// no InputVarsOptions.
func InputStreamNestedObject(varsAreSensitive bool, useMarkdownDescription bool) schema.NestedAttributeObject {
	enabledDescription, enabledMarkdown := splitDescription("Enable the stream.", useMarkdownDescription)
	conditionDescription, conditionMarkdown := splitDescription(
		"Agent condition expression to evaluate whether to apply this stream.", useMarkdownDescription)
	varsDescription, varsMarkdown := splitDescription("Stream-level variables as JSON.", useMarkdownDescription)

	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			AttrEnabled: schema.BoolAttribute{
				Description:         enabledDescription,
				MarkdownDescription: enabledMarkdown,
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			AttrCondition: schema.StringAttribute{
				Description:         conditionDescription,
				MarkdownDescription: conditionMarkdown,
				Optional:            true,
			},
			AttrVars: schema.StringAttribute{
				Description:         varsDescription,
				MarkdownDescription: varsMarkdown,
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Sensitive:           varsAreSensitive,
			},
		},
	}
}
