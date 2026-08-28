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
//
// This is a thin, logging-aware wrapper around
// elasticsearch.DeleteWithNotFoundAsSuccess, which owns the actual
// 404-as-success classification and diagnostic construction so the
// convention has a single implementation shared by both layers.
func DeleteWithNotFoundAsSuccess(ctx context.Context, kindLabel, id string, do func() error) fwdiags.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("Deleting %s: %s", kindLabel, id))

	err := do()
	switch {
	case err == nil:
		tflog.Debug(ctx, fmt.Sprintf("Successfully deleted %s: %s", kindLabel, id))
	case elasticsearch.IsNotFoundElasticsearchError(err):
		tflog.Debug(ctx, fmt.Sprintf("%s already absent: %s", kindLabel, id))
	default:
		err = fmt.Errorf("unable to delete %s: %s — %w", kindLabel, id, err)
	}

	return elasticsearch.DeleteWithNotFoundAsSuccess(err, fmt.Sprintf("Failed to delete %s", kindLabel))
}

// RequireNonEmptyID returns diagnostics with a single error when id is
// empty, using fieldName (e.g. "calendar_id") in the message. Callers should
// return their own zero-value result together with these diagnostics when
// diags.HasError() is true.
func RequireNonEmptyID(id, fieldName string) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics
	if id == "" {
		diags.AddError("Invalid resource ID", fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return diags
}

// ReadWithNotFoundAsAbsent runs do to fetch an ML sub-resource. A "not found"
// error from Elasticsearch is treated as the resource being absent
// (found=false, no diagnostics) rather than an error, matching Terraform's
// refresh-time drift-detection convention. kindLabel and id are used to build
// consistent error messages across the ML sub-resource packages.
//
// This is a thin, logging-aware wrapper around elasticsearch.CallOrNotFound,
// which owns the actual 404-as-absent classification and diagnostic
// construction so the convention has a single implementation shared by both
// layers.
func ReadWithNotFoundAsAbsent[T any](ctx context.Context, kindLabel, id string, do func() (T, error)) (result T, found bool, diags fwdiags.Diagnostics) {
	tflog.Debug(ctx, fmt.Sprintf("Reading %s: %s", kindLabel, id))

	var notFound bool
	result, diags = elasticsearch.CallOrNotFound(func() (T, error) {
		res, err := do()
		switch {
		case err == nil:
			return res, nil
		case elasticsearch.IsNotFoundElasticsearchError(err):
			notFound = true
			return res, err
		default:
			return res, fmt.Errorf("unable to get %s: %s — %w", kindLabel, id, err)
		}
	}, fmt.Sprintf("Failed to get %s", kindLabel))

	if diags.HasError() {
		return result, false, diags
	}
	return result, !notFound, diags
}
