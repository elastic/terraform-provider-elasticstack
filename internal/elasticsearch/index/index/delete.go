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

package index

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// deleteIndex is the envelope delete callback. It checks deletion_protection
// before calling the Delete Index API.
func deleteIndex(ctx context.Context, client *clients.ElasticsearchScopedClient, resourceID string, model tfModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if model.DeletionProtection.ValueBool() {
		diags.AddAttributeError(
			path.Root("deletion_protection"),
			"cannot destroy index without setting deletion_protection=false and running `terraform apply`",
			"cannot destroy index without setting deletion_protection=false and running `terraform apply`",
		)
		return diags
	}

	diags.Append(elasticsearch.DeleteIndex(ctx, client, resourceID)...)
	return diags
}
