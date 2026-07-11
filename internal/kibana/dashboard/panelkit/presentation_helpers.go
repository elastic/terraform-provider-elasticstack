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
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ApplyPresentationFromAPI updates the four standard optional presentation fields in Terraform state
// (title, description, hide_title, hide_border) from API pointer values, using null-preservation
// semantics (REQ-009): if a field is already null/unknown in state it stays unchanged.
func ApplyPresentationFromAPI(
	existingTitle *types.String,
	existingDescription *types.String,
	existingHideTitle *types.Bool,
	existingHideBorder *types.Bool,
	apiTitle *string,
	apiDescription *string,
	apiHideTitle *bool,
	apiHideBorder *bool,
) {
	*existingTitle = PreserveString(*existingTitle, apiTitle)
	*existingDescription = PreserveString(*existingDescription, apiDescription)
	*existingHideTitle = PreserveBool(*existingHideTitle, apiHideTitle)
	*existingHideBorder = PreserveBool(*existingHideBorder, apiHideBorder)
}

// NullPreservePresentationFromPrior applies null-intent preservation for the four standard
// presentation fields (title, description, hide_title, hide_border): if a field was null/unknown
// in the prior state, it is reset to null in existing.
func NullPreservePresentationFromPrior(
	priorTitle types.String,
	priorDescription types.String,
	priorHideTitle types.Bool,
	priorHideBorder types.Bool,
	existingTitle *types.String,
	existingDescription *types.String,
	existingHideTitle *types.Bool,
	existingHideBorder *types.Bool,
) {
	if !typeutils.IsKnown(priorTitle) {
		*existingTitle = types.StringNull()
	}
	if !typeutils.IsKnown(priorDescription) {
		*existingDescription = types.StringNull()
	}
	if !typeutils.IsKnown(priorHideTitle) {
		*existingHideTitle = types.BoolNull()
	}
	if !typeutils.IsKnown(priorHideBorder) {
		*existingHideBorder = types.BoolNull()
	}
}

// NullPreserveBaseFromPrior is a nil-safe wrapper around NullPreservePresentationFromPrior.
// It is a no-op when any existing pointer is nil, allowing panel helpers to safely call it
// without a separate nil-guard on the four standard presentation fields.
func NullPreserveBaseFromPrior(
	priorTitle types.String,
	priorDescription types.String,
	priorHideTitle types.Bool,
	priorHideBorder types.Bool,
	existingTitle *types.String,
	existingDescription *types.String,
	existingHideTitle *types.Bool,
	existingHideBorder *types.Bool,
) {
	if existingTitle == nil || existingDescription == nil || existingHideTitle == nil || existingHideBorder == nil {
		return
	}
	NullPreservePresentationFromPrior(
		priorTitle, priorDescription, priorHideTitle, priorHideBorder,
		existingTitle, existingDescription, existingHideTitle, existingHideBorder,
	)
}

// BuildPresentationConfig writes the four standard optional presentation fields from Terraform state
// into API pointer fields (title, description, hide_title, hide_border).
func BuildPresentationConfig(
	cfgTitle types.String,
	cfgDescription types.String,
	cfgHideTitle types.Bool,
	cfgHideBorder types.Bool,
	apiTitle **string,
	apiDescription **string,
	apiHideTitle **bool,
	apiHideBorder **bool,
) {
	if typeutils.IsKnown(cfgTitle) {
		*apiTitle = cfgTitle.ValueStringPointer()
	}
	if typeutils.IsKnown(cfgDescription) {
		*apiDescription = cfgDescription.ValueStringPointer()
	}
	if typeutils.IsKnown(cfgHideTitle) {
		*apiHideTitle = cfgHideTitle.ValueBoolPointer()
	}
	if typeutils.IsKnown(cfgHideBorder) {
		*apiHideBorder = cfgHideBorder.ValueBoolPointer()
	}
}
