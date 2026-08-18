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

package kibanaoapi

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// GetWithNamespaceRetry calls apiGet with params, and if the lookup returns
// not-found (a nil result with no error) while namespaceTypeKnown is false,
// retries once after calling setAgnosticNamespaceType to widen params to
// namespace_type=agnostic. Security exception list/item reads use this so
// that import - where the namespace type is not yet known - still finds
// resources created with a non-default namespace type.
func GetWithNamespaceRetry[P any, R any](
	ctx context.Context,
	client *Client,
	spaceID string,
	params *P,
	namespaceTypeKnown bool,
	setAgnosticNamespaceType func(params *P),
	apiGet func(ctx context.Context, client *Client, spaceID string, params *P) (*R, diag.Diagnostics),
) (*R, diag.Diagnostics) {
	data, diags := apiGet(ctx, client, spaceID, params)
	if diags.HasError() || data != nil || namespaceTypeKnown {
		return data, diags
	}

	setAgnosticNamespaceType(params)
	return apiGet(ctx, client, spaceID, params)
}
