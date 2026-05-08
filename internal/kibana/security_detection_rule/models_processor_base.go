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

package securitydetectionrule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// baseRuleProcessor provides the three interface methods that are structurally identical
// across all rule processor types. Embed it in each concrete processor and supply the
// type-specific updateFn and idFn at construction time via the concrete processor's
// constructor (e.g. newEqlRuleProcessor).
type baseRuleProcessor[T any] struct {
	updateFn func(ctx context.Context, v *T, d *Data) diag.Diagnostics
	idFn     func(v T) string
}

func (b baseRuleProcessor[T]) HandlesAPIRuleResponse(rule any) bool {
	return handlesAPIRuleResponse[T](rule)
}

func (b baseRuleProcessor[T]) UpdateFromResponse(ctx context.Context, rule any, d *Data) diag.Diagnostics {
	return updateFromRuleResponse[T](rule, func(v *T) diag.Diagnostics {
		return b.updateFn(ctx, v, d)
	})
}

func (b baseRuleProcessor[T]) ExtractID(response any) (string, diag.Diagnostics) {
	return extractRuleID[T](response, b.idFn)
}
