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

package systemuser

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// deleteSystemUser does not make an API call because system users are built-in
// and cannot be deleted via the Elasticsearch API. We simply remove the resource
// from Terraform state.
func deleteSystemUser(ctx context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, _ Data) diag.Diagnostics {
	tflog.Warn(ctx, fmt.Sprintf(`System user '%s' is not deletable, just removing from state`, resourceID))
	return nil
}
