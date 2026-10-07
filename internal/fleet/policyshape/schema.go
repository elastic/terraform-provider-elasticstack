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

// InputVarsOptions configures the input-level "vars" attribute, the one
// attribute that intentionally differs in shape between callers.
type InputVarsOptions struct {
	// Computed marks "vars" Computed and adds stringplanmodifier.UseStateForUnknown.
	// managedintegration needs this because some packages (e.g.
	// cloud_security_posture/CSPM) populate informational input-level vars
	// (such as CloudFormation quick-create template URLs) that are always
	// present in the API response regardless of configuration; Computed lets
	// those flow through without requiring the user to declare them.
	// integration_policy's vars attribute is purely user-supplied.
	Computed bool
	// Description is this caller's wording for the input-level "vars"
	// attribute (the two callers' text differs).
	Description string
}

// InputsNestedObjectConfig carries the resource-specific pieces that
// InputsNestedObject needs to serve both integration_policy and
// managedintegration. All descriptions are set as MarkdownDescription.
type InputsNestedObjectConfig struct {
	// CustomType is the nested object's CustomType. Each caller wraps a
	// different attribute-types map in its own InputType (integration_policy
	// includes "defaults"; managedintegration doesn't), so this is supplied
	// by the caller rather than built here.
	CustomType basetypes.ObjectTypable
	// VarsAreSensitive marks every "vars" attribute (input- and
	// stream-level) Sensitive.
	VarsAreSensitive bool
	// IncludeDefaults adds the package-computed "defaults" sub-object nested
	// under each input (integration_policy only).
	IncludeDefaults bool
	// InputVars configures the input-level "vars" attribute.
	InputVars InputVarsOptions
}

// InputsNestedObject builds the "inputs" map's NestedAttributeObject shared
// by integration_policy and managedintegration: enabled/condition/vars/
// streams, plus an optional package-computed "defaults" sub-object. See
// InputsNestedObjectConfig for the pieces that intentionally differ between
// those two callers.
func InputsNestedObject(cfg InputsNestedObjectConfig) schema.NestedAttributeObject {
	varsAttribute := schema.StringAttribute{
		MarkdownDescription: cfg.InputVars.Description,
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
			MarkdownDescription: "Enable the input.",
			Computed:            true,
			Optional:            true,
			Default:             booldefault.StaticBool(true),
		},
		AttrCondition: schema.StringAttribute{
			MarkdownDescription: "Agent condition expression to evaluate whether to apply this input.",
			Optional:            true,
		},
		AttrVars: varsAttribute,
		AttrStreams: schema.MapNestedAttribute{
			MarkdownDescription: "Input streams mapped by stream ID.",
			Optional:            true,
			NestedObject:        InputStreamNestedObject(cfg.VarsAreSensitive),
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
		MarkdownDescription: "Input defaults.",
		Computed:            true,
		Default: objectdefault.StaticValue(basetypes.NewObjectNull(
			InputDefaultsAttributeTypes(),
		)),
		Attributes: map[string]schema.Attribute{
			AttrVars: schema.StringAttribute{
				MarkdownDescription: "Input-level variable defaults as JSON.",
				CustomType:          jsontypes.NormalizedType{},
				Computed:            true,
			},
			AttrStreams: schema.MapNestedAttribute{
				MarkdownDescription: "Stream-level defaults mapped by stream ID.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						AttrEnabled: schema.BoolAttribute{
							MarkdownDescription: "Default enabled state for the stream.",
							Computed:            true,
						},
						AttrVars: schema.StringAttribute{
							MarkdownDescription: "Stream-level variable defaults as JSON.",
							CustomType:          jsontypes.NormalizedType{},
							Computed:            true,
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
func InputStreamNestedObject(varsAreSensitive bool) schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			AttrEnabled: schema.BoolAttribute{
				MarkdownDescription: "Enable the stream.",
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(true),
			},
			AttrCondition: schema.StringAttribute{
				MarkdownDescription: "Agent condition expression to evaluate whether to apply this stream.",
				Optional:            true,
			},
			AttrVars: schema.StringAttribute{
				MarkdownDescription: "Stream-level variables as JSON.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Sensitive:           varsAreSensitive,
			},
		},
	}
}
