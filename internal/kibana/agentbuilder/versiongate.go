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

// Package agentbuilder provides shared helpers for Agent Builder resource packages.
package agentbuilder

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// EnforceVersion checks that the Kibana server meets minVersion for Agent Builder
// entities. It appends any diagnostics to diags and returns false if the check
// fails or the version is not met.
func EnforceVersion(ctx context.Context, client *clients.KibanaScopedClient, minVersion *version.Version, entityName string, diags *fwdiags.Diagnostics) bool {
	supported, sdkDiags := client.EnforceMinVersion(ctx, minVersion)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return false
	}
	if !supported {
		diags.AddError("Unsupported server version",
			fmt.Sprintf("Agent Builder %s require Elastic Stack v%s or later.", entityName, minVersion))
		return false
	}
	return true
}
