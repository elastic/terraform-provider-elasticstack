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

package entitycore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// WithVersionRequirements is an optional interface that Kibana entity models
// may implement to declare server version requirements. When a decoded model
// satisfies this interface, the generic Kibana envelopes evaluate the
// requirements after scoped client resolution and before invoking the concrete
// lifecycle callback.
type WithVersionRequirements interface {
	GetVersionRequirements() ([]DataSourceVersionRequirement, diag.Diagnostics)
}

// enforceVersionRequirements checks whether model implements
// WithVersionRequirements and, if so, evaluates each requirement against the
// scoped client. It returns any diagnostics produced.
func enforceVersionRequirements(ctx context.Context, client *clients.KibanaScopedClient, model any) diag.Diagnostics {
	var diags diag.Diagnostics
	versionModel, ok := model.(WithVersionRequirements)
	if !ok {
		return diags
	}

	reqs, vDiags := versionModel.GetVersionRequirements()
	diags.Append(vDiags...)
	if diags.HasError() {
		return diags
	}

	for _, vReq := range reqs {
		supported, sdkDiags := client.EnforceMinVersion(ctx, &vReq.MinVersion)
		diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
		if diags.HasError() {
			return diags
		}
		if !supported {
			diags.AddError("Unsupported server version", vReq.ErrorMessage)
			return diags
		}
	}

	return diags
}
