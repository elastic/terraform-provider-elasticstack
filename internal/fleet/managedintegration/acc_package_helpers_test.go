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

package managedintegration_test

import (
	"context"
	"os"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

const cspmPackageName = "cloud_security_posture"

// cspmVersionPairCandidates lists preferred (from, to) semver pairs for
// in-place package.version acceptance tests. Pairs are tried in order; the
// first pair whose versions are known to Fleet and can be installed wins.
var cspmVersionPairCandidates = [][2]string{
	{"3.4.0", "3.5.0"},
	{"3.3.0", "3.4.0"},
}

// resolveCSPMInPlaceVersionUpgradePair returns two distinct cloud_security_posture
// package versions suitable for an in-place package.version update test. It
// verifies each version against the live Fleet registry and ensures both are
// installed before returning. Skips the test when no pair is available.
func resolveCSPMInPlaceVersionUpgradePair(t *testing.T) (fromVersion, toVersion string) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping: TF_ACC not set")
	}

	fc := mustFleetClient(t)
	ctx := context.Background()

	for _, pair := range cspmVersionPairCandidates {
		from, to := pair[0], pair[1]
		if !cspmPackageVersionKnown(ctx, fc, from) || !cspmPackageVersionKnown(ctx, fc, to) {
			continue
		}
		ensureCSPMPackageInstalled(t, fc, from)
		ensureCSPMPackageInstalled(t, fc, to)
		return from, to
	}

	from, to := cspmInstalledAndLatestPair(ctx, t, fc)
	if from == "" || to == "" || from == to {
		t.Skip("skipping: need two distinct installable cloud_security_posture package versions for in-place version update test")
	}
	ensureCSPMPackageInstalled(t, fc, from)
	ensureCSPMPackageInstalled(t, fc, to)
	return from, to
}

func cspmPackageVersionKnown(ctx context.Context, fc *fleetclient.Client, version string) bool {
	pkg, diags := fleetclient.GetPackage(ctx, fc, cspmPackageName, version, "default")
	return !diags.HasError() && pkg != nil
}

func ensureCSPMPackageInstalled(t *testing.T, fc *fleetclient.Client, version string) {
	t.Helper()
	diags := fleetclient.InstallPackage(t.Context(), fc, cspmPackageName, version, fleetclient.InstallPackageOptions{
		Force: true,
	})
	if diags.HasError() {
		t.Fatalf("failed to install %s/%s for acceptance test: %v", cspmPackageName, version, diags)
	}
}

// cspmInstalledAndLatestPair uses the Fleet package list to find the currently
// installed CSPM version and a newer registry version when candidate pairs are
// unavailable.
func cspmInstalledAndLatestPair(ctx context.Context, t *testing.T, fc *fleetclient.Client) (from, to string) {
	t.Helper()
	items, diags := fleetclient.GetPackages(ctx, fc, false, "default")
	if diags.HasError() {
		t.Skipf("skipping: could not list Fleet packages: %v", diags)
		return "", ""
	}
	for _, item := range items {
		if item.Name != cspmPackageName {
			continue
		}
		installed := item.Version
		if item.InstallationInfo != nil && item.InstallationInfo.Version != "" {
			installed = item.InstallationInfo.Version
		}
		candidate := item.Version
		if item.LatestVersion != nil && *item.LatestVersion != "" {
			candidate = *item.LatestVersion
		}
		if installed == "" || candidate == "" {
			return "", ""
		}
		installedVer, err1 := semver.NewVersion(installed)
		candidateVer, err2 := semver.NewVersion(candidate)
		if err1 != nil || err2 != nil || !candidateVer.GreaterThan(installedVer) {
			return "", ""
		}
		if !cspmPackageVersionKnown(ctx, fc, installed) || !cspmPackageVersionKnown(ctx, fc, candidate) {
			return "", ""
		}
		return installed, candidate
	}
	return "", ""
}
