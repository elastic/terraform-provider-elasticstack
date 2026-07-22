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
	"strings"

	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/go-version"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

// versionFetcher is a function that retrieves the server's raw version string
// and build flavor. It is used by the shared enforceMinVersion and
// enforceVersionCheck helpers to decouple the fetch mechanism from the
// version-constraint logic.
type versionFetcher func(ctx context.Context) (rawVersion, flavor string, diags fwdiag.Diagnostics)

// minVersionCheck compares a parsed server version against a release minimum.
type minVersionCheck func(server, minimum *version.Version) bool

// versionAtLeastRelease reports whether server satisfies minimum using strict
// semver GreaterThanOrEqual (used for Elasticsearch EnforceMinVersion).
func versionAtLeastRelease(server, minimum *version.Version) bool {
	return server.GreaterThanOrEqual(minimum)
}

// kibanaVersionAtLeastRelease reports whether a Kibana server version satisfies
// a release minimum. It uses strict semver first, then applies the Elastic
// Kibana/CI convention that a same-core -SNAPSHOT build satisfies a release
// floor (e.g. 9.5.0-SNAPSHOT meets minimum 9.5.0). Elasticsearch
// EnforceMinVersion does not apply this uplift.
func kibanaVersionAtLeastRelease(server, minimum *version.Version) bool {
	if server.GreaterThanOrEqual(minimum) {
		return true
	}
	return snapshotBuildSatisfiesReleaseMinimum(server, minimum)
}

// snapshotBuildSatisfiesReleaseMinimum is true when minimum is a release version,
// server shares the same core version, and server's prerelease is SNAPSHOT.
// Other prereleases (e.g. beta) do not satisfy a release minimum.
func snapshotBuildSatisfiesReleaseMinimum(server, minimum *version.Version) bool {
	if minimum.Prerelease() != "" || server.Prerelease() == "" {
		return false
	}
	if !server.Core().Equal(minimum.Core()) {
		return false
	}
	return strings.EqualFold(server.Prerelease(), "SNAPSHOT")
}

// applyVersionConstraint evaluates check against rawVersion, short-circuiting
// to true for serverless clusters. It is the shared core of EnforceMinVersion
// and EnforceVersionCheck for both ElasticsearchScopedClient and KibanaScopedClient.
func applyVersionConstraint(
	flavor, rawVersion string,
	check func(*version.Version) bool,
) (bool, fwdiag.Diagnostics) {
	if flavor == ServerlessFlavor {
		return true, nil
	}
	sv, err := version.NewVersion(rawVersion)
	if err != nil {
		return false, diagutil.FrameworkDiagFromError(err)
	}
	return check(sv), nil
}

// enforceMinVersion implements the shared body of EnforceMinVersion for both
// scoped client types. It short-circuits to true when minVersion is nil, then
// delegates to the provided fetch function to obtain the server version, and
// finally applies meets via applyVersionConstraint.
func enforceMinVersion(ctx context.Context, minVersion *version.Version, fetch versionFetcher, meets minVersionCheck) (bool, fwdiag.Diagnostics) {
	if minVersion == nil {
		return true, nil
	}
	rawVersion, flavor, diags := fetch(ctx)
	if diags.HasError() {
		return false, diags
	}
	return applyVersionConstraint(flavor, rawVersion, func(sv *version.Version) bool {
		return meets(sv, minVersion)
	})
}

// enforceVersionCheck implements the shared body of EnforceVersionCheck for both
// scoped client types. It delegates to the provided fetch function to obtain the
// server version and then applies the caller-supplied check via applyVersionConstraint.
func enforceVersionCheck(ctx context.Context, check func(*version.Version) bool, fetch versionFetcher) (bool, fwdiag.Diagnostics) {
	rawVersion, flavor, diags := fetch(ctx)
	if diags.HasError() {
		return false, diags
	}
	return applyVersionConstraint(flavor, rawVersion, check)
}
