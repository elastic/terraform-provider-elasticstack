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

package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PanelConfigBlockOpts configures PanelConfigBlock.
type PanelConfigBlockOpts struct {
	// Description is the base markdown description for the typed config block. PanelConfigBlock
	// extends it with the standard "Mutually exclusive with ..." sibling clause.
	Description string
	// BlockName is the typed sibling attribute name (e.g. "slo_burn_rate_config"). Used to derive
	// ConflictsWith expressions and the sibling list in the description.
	BlockName string
	// PanelType is the Kibana API panel discriminator (e.g. "slo_burn_rate"). Used by
	// AllowedIf / RequiredIf validators that scope the block to its panel type.
	PanelType string
	// Required reports whether the block is required when the panel type matches. Defaults to false
	// (optional even when the type matches, e.g. control panels with omitted typed config).
	Required bool
	// Attributes is the typed block's attribute set.
	Attributes map[string]schema.Attribute
	// ExtraValidators are appended after the standard ConflictsWith + AllowedIf (+ RequiredIf) set.
	ExtraValidators []validator.Object
}

// PanelConfigBlock returns the standard SingleNestedAttribute scaffold every panel handler emits:
// markdown description with sibling-block exclusion clause, the optional flag, the typed sibling
// ConflictsWith expressions, the AllowedIf/RequiredIf panel-type guard, plus any panel-specific
// validators. Handlers compose by passing their `Attributes` map and panel-specific options.
func PanelConfigBlock(opts PanelConfigBlockOpts) schema.Attribute {
	siblings := TypedSiblingPanelConfigBlockNames()
	typePath := path.MatchRelative().AtParent().AtName("type")
	panelTypes := []string{opts.PanelType}

	vs := []validator.Object{
		objectvalidator.ConflictsWith(SiblingTypedPanelConfigConflictPathsExcept(opts.BlockName, siblings)...),
		validators.AllowedIfDependentPathExpressionOneOf(typePath, panelTypes),
	}
	if opts.Required {
		vs = append(vs, validators.RequiredIfDependentPathExpressionOneOf(typePath, panelTypes))
	}
	vs = append(vs, opts.ExtraValidators...)

	return schema.SingleNestedAttribute{
		MarkdownDescription: PanelConfigDescription(opts.Description, opts.BlockName, siblings),
		Optional:            true,
		Attributes:          opts.Attributes,
		Validators:          vs,
	}
}
