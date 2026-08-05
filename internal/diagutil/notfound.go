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

package diagutil

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// WarnNotFoundAndKeepState logs a warning that the kind identified by id was
// not found upstream and returns state unchanged with found=false, matching
// the "not found -> warn -> early-return" envelope shape shared by the
// ElasticsearchScopedClient-backed read callbacks (signature
// (Data, bool, diag.Diagnostics)). diags is returned unmodified so callers
// keep any diagnostics already accumulated before the not-found check.
func WarnNotFoundAndKeepState[T any](ctx context.Context, kind, id string, state T, diags diag.Diagnostics) (T, bool, diag.Diagnostics) {
	tflog.Warn(ctx, fmt.Sprintf(`%s "%s" not found, removing from state`, kind, id))
	return state, false, diags
}
