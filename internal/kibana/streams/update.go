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

package streams

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func updateStream(ctx context.Context, client *clients.KibanaScopedClient, _, _ string, plan, _ streamModel) (streamModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	readModel, upsertDiags := upsertStream(ctx, client, plan)
	diags.Append(upsertDiags...)
	if diags.HasError() {
		return streamModel{}, diags
	}
	if readModel == nil {
		diags.AddError("Error reading stream after update", "The stream was updated but could not be read back.")
		return streamModel{}, diags
	}

	return *readModel, diags
}
