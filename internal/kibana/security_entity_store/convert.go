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

package security_entity_store

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// stringSliceToAPITypes converts a []string to a pointer to a slice of a ~string
// type alias, covering the pattern used by generated kbapi enum slices.
func stringSliceToAPITypes[T ~string](values []string) *[]T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, 0, len(values))
	for _, v := range values {
		out = append(out, T(v))
	}
	return &out
}

// apiLogExtraction matches the generated Kibana API log extraction body shape.
type apiLogExtraction[T ~string] = struct {
	AdditionalIndexPatterns     *[]string `json:"additionalIndexPatterns,omitempty"`
	Delay                       *string   `json:"delay,omitempty"`
	DocsLimit                   *int      `json:"docsLimit,omitempty"`
	ExcludedIndexPatterns       *[]string `json:"excludedIndexPatterns,omitempty"`
	FieldHistoryLength          *int      `json:"fieldHistoryLength,omitempty"`
	Frequency                   *string   `json:"frequency,omitempty"`
	LookbackPeriod              *string   `json:"lookbackPeriod,omitempty"`
	MaxLogsPerPage              *int      `json:"maxLogsPerPage,omitempty"`
	MaxLogsPerWindow            *int      `json:"maxLogsPerWindow,omitempty"`
	MaxLogsPerWindowCapBehavior *T        `json:"maxLogsPerWindowCapBehavior,omitempty"`
	MaxTimeWindowSize           *string   `json:"maxTimeWindowSize,omitempty"`
}

func expandLogExtractionCommon[T ~string](ctx context.Context, obj types.Object) (*apiLogExtraction[T], diag.Diagnostics) {
	var model logExtractionModel
	var diags diag.Diagnostics
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}
	add := typeutils.CollectionToSliceStringPtr(ctx, model.AdditionalIndexPatterns, path.Empty(), &diags)
	excl := typeutils.CollectionToSliceStringPtr(ctx, model.ExcludedIndexPatterns, path.Empty(), &diags)
	if diags.HasError() {
		return nil, diags
	}
	c := &apiLogExtraction[T]{
		AdditionalIndexPatterns: add,
		Delay:                   typeutils.OptionalString(model.Delay),
		DocsLimit:               typeutils.OptionalInt(model.DocsLimit),
		ExcludedIndexPatterns:   excl,
		FieldHistoryLength:      typeutils.OptionalInt(model.FieldHistoryLength),
		Frequency:               typeutils.OptionalString(model.Frequency),
		LookbackPeriod:          typeutils.OptionalString(model.LookbackPeriod),
		MaxLogsPerPage:          typeutils.OptionalInt(model.MaxLogsPerPage),
		MaxLogsPerWindow:        typeutils.OptionalInt(model.MaxLogsPerWindow),
		MaxTimeWindowSize:       typeutils.OptionalString(model.MaxTimeWindowSize),
	}
	if typeutils.IsKnown(model.MaxLogsPerWindowCapBehavior) {
		behavior := T(model.MaxLogsPerWindowCapBehavior.ValueString())
		c.MaxLogsPerWindowCapBehavior = &behavior
	}
	return c, diags
}
