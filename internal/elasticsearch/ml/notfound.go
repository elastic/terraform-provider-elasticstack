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

package ml

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// DeleteWithNotFoundAsSuccess runs do to delete an ML sub-resource, treating
// a "not found" error from Elasticsearch as a successful no-op (the resource
// is already gone). kindLabel (e.g. "ML filter") and id are used to build
// consistent log and error messages across the ML sub-resource packages.
func DeleteWithNotFoundAsSuccess(ctx context.Context, kindLabel, id string, do func() error) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	tflog.Debug(ctx, fmt.Sprintf("Deleting %s: %s", kindLabel, id))

	if err := do(); err != nil {
		if elasticsearch.IsNotFoundElasticsearchError(err) {
			tflog.Debug(ctx, fmt.Sprintf("%s already absent: %s", kindLabel, id))
			return diags
		}
		diags.AddError(fmt.Sprintf("Failed to delete %s", kindLabel), fmt.Sprintf("Unable to delete %s: %s — %s", kindLabel, id, err.Error()))
		return diags
	}

	tflog.Debug(ctx, fmt.Sprintf("Successfully deleted %s: %s", kindLabel, id))
	return diags
}

// ReadWithNotFoundAsAbsent runs do to fetch an ML sub-resource. A "not found"
// error from Elasticsearch is treated as the resource being absent
// (found=false, no diagnostics) rather than an error, matching Terraform's
// refresh-time drift-detection convention. kindLabel and id are used to build
// consistent error messages across the ML sub-resource packages.
func ReadWithNotFoundAsAbsent[T any](ctx context.Context, kindLabel, id string, do func() (T, error)) (result T, found bool, diags fwdiags.Diagnostics) {
	tflog.Debug(ctx, fmt.Sprintf("Reading %s: %s", kindLabel, id))

	result, err := do()
	if err != nil {
		if elasticsearch.IsNotFoundElasticsearchError(err) {
			return result, false, diags
		}
		diags.AddError(fmt.Sprintf("Failed to get %s", kindLabel), fmt.Sprintf("Unable to get %s: %s — %s", kindLabel, id, err.Error()))
		return result, false, diags
	}

	return result, true, diags
}
