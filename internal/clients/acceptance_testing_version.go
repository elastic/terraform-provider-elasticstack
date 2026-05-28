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

package clients

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// AcceptanceServerInfo returns the connected Elasticsearch server version and a boolean
// indicating whether the cluster is serverless. It is exposed only for acceptance-test skip
// plumbing in internal/versionutils. Production code SHALL NOT call this.
func AcceptanceServerInfo(ctx context.Context, c *ElasticsearchScopedClient) (*version.Version, bool, fwdiag.Diagnostics) {
	info, diags := c.serverInfo(ctx)
	if diags.HasError() {
		return nil, false, diags
	}

	serverVersion, err := version.NewVersion(info.Version.Int)
	if err != nil {
		return nil, false, diagutil.FrameworkDiagFromError(err)
	}

	return serverVersion, info.Version.BuildFlavor == ServerlessFlavor, nil
}
