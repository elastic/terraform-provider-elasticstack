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

package templateutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TemplateModel is implemented by plan/state models that embed a template block.
// WithTemplate must return a copy of the receiver with Template replaced.
type TemplateModel[T any] interface {
	GetTemplate() types.Object
	WithTemplate(tpl types.Object) T
}

// ReconcilePlanModelForSemanticDrift is a shared wrapper around
// ReconcileTemplateWithPriorStateForSemanticDrift for full plan/state model types.
// It extracts the Template field, reconciles it, and returns an updated model copy when
// changes are needed. Returns (nil, nil diags) when no reconciliation is required.
func ReconcilePlanModelForSemanticDrift[M TemplateModel[M]](
	ctx context.Context,
	plan, state, config M,
	attrTypes func() map[string]attr.Type,
) (*M, diag.Diagnostics) {
	newTpl, changed, diags := ReconcileTemplateWithPriorStateForSemanticDrift(
		ctx,
		plan.GetTemplate(),
		state.GetTemplate(),
		config.GetTemplate(),
		attrTypes(),
	)
	if diags.HasError() || !changed {
		return nil, diags
	}
	out := plan.WithTemplate(newTpl)
	return &out, diags
}
