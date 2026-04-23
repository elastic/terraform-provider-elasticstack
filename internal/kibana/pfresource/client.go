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

package pfresource

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type VersionRequirement struct {
	MinimumVersion *version.Version
	ErrorSummary   string
	ErrorDetail    string
}

func ResolveKibanaClient(ctx context.Context, factory *clients.ProviderClientFactory, kibanaConnection types.List) (*clients.KibanaScopedClient, diag.Diagnostics) {
	if factory == nil {
		return nil, diag.Diagnostics{diag.NewErrorDiagnostic(
			"Provider not configured",
			"Expected configured provider client factory. Please report this issue to the provider developers.",
		)}
	}
	return factory.GetKibanaClient(ctx, kibanaConnection)
}

func EnforceVersion(ctx context.Context, client *clients.KibanaScopedClient, requirement VersionRequirement) diag.Diagnostics {
	if requirement.MinimumVersion == nil {
		return nil
	}

	supported, sdkDiags := client.EnforceMinVersion(ctx, requirement.MinimumVersion)
	fwDiags := diagutil.FrameworkDiagsFromSDK(sdkDiags)
	if fwDiags.HasError() {
		return fwDiags
	}
	if !supported {
		return diag.Diagnostics{diag.NewErrorDiagnostic(requirement.ErrorSummary, requirement.ErrorDetail)}
	}
	return nil
}
