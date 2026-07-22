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
	"fmt"
	"os"
	"sync"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
)

const cspmPackageName = "cloud_security_posture"

// cspmVersionPairCandidates lists preferred (from, to) semver pairs for
// in-place package.version acceptance tests. Only adjacent, same-major bumps are
// considered safe for live upgrade tests.
var cspmVersionPairCandidates = [][2]string{
	{"3.4.0", "3.5.0"},
	{"3.3.0", "3.4.0"},
}

var (
	cspmPinnedPrecheckOnce   sync.Once
	cspmPinnedPrecheckOK     bool
	cspmPinnedPrecheckReason string
)

// skipUnlessCSPMPinnedPackageAvailable ensures cloud_security_posture/cspmPackageVersion
// is known to Fleet and installed once per test process without repeated force installs.
func skipUnlessCSPMPinnedPackageAvailable(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		return
	}
	cspmPinnedPrecheckOnce.Do(func() {
		fc, err := acceptanceFleetClient()
		if err != nil {
			cspmPinnedPrecheckReason = err.Error()
			return
		}
		ctx := context.Background()
		if !cspmPackageVersionKnown(ctx, fc, cspmPackageVersion) {
			cspmPinnedPrecheckReason = fmt.Sprintf("%s/%s is not available in the Fleet registry", cspmPackageName, cspmPackageVersion)
			return
		}
		if cspmPackageInstalledInDefaultSpace(ctx, fc, cspmPackageVersion) {
			cspmPinnedPrecheckOK = true
			return
		}
		if err := tryInstallCSPMPackage(ctx, fc, cspmPackageVersion); err != nil {
			cspmPinnedPrecheckReason = err.Error()
			return
		}
		cspmPinnedPrecheckOK = true
	})
	if !cspmPinnedPrecheckOK {
		if cspmPinnedPrecheckReason == "" {
			cspmPinnedPrecheckReason = fmt.Sprintf("%s/%s could not be prepared for acceptance tests", cspmPackageName, cspmPackageVersion)
		}
		t.Skip("skipping: " + cspmPinnedPrecheckReason)
	}
}

func acceptanceFleetClient() (*fleetclient.Client, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, fmt.Errorf("could not create Kibana client for Fleet package precheck: %w", err)
	}
	return client.GetFleetClient(), nil
}

func cspmPackageInstalledInDefaultSpace(ctx context.Context, fc *fleetclient.Client, version string) bool {
	pkg, diags := fleetclient.GetPackage(ctx, fc, cspmPackageName, version, managedIntegrationDefaultSpace)
	if diags.HasError() || pkg == nil || pkg.InstallationInfo == nil {
		return false
	}
	return pkg.InstallationInfo.InstallStatus == kbapi.KibanaHTTPAPIsPackageInfoInstallationInfoInstallStatusInstalled
}

func tryInstallCSPMPackage(ctx context.Context, fc *fleetclient.Client, version string) error {
	diags := fleetclient.InstallPackage(ctx, fc, cspmPackageName, version, fleetclient.InstallPackageOptions{})
	if diags.HasError() {
		return fmt.Errorf("install %s/%s: %v", cspmPackageName, version, diags)
	}
	return nil
}

// resolveCSPMInPlaceVersionUpgradePair returns two distinct, safely adjacent
// cloud_security_posture versions for an in-place package.version update test.
func resolveCSPMInPlaceVersionUpgradePair(t *testing.T) (fromVersion, toVersion string) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("skipping: TF_ACC not set")
	}

	fc, err := acceptanceFleetClient()
	if err != nil {
		t.Skip("skipping: " + err.Error())
	}
	ctx := context.Background()

	for _, pair := range cspmVersionPairCandidates {
		from, to := pair[0], pair[1]
		if !isSafeAdjacentCSPMUpgrade(from, to) {
			continue
		}
		if !cspmPackageVersionKnown(ctx, fc, from) || !cspmPackageVersionKnown(ctx, fc, to) {
			continue
		}
		if err := ensureCSPMPackageInstalledOptional(ctx, fc, from); err != nil {
			t.Skipf("skipping: cannot prepare %s/%s for version update test: %v", cspmPackageName, from, err)
		}
		if err := ensureCSPMPackageInstalledOptional(ctx, fc, to); err != nil {
			t.Skipf("skipping: cannot prepare %s/%s for version update test: %v", cspmPackageName, to, err)
		}
		return from, to
	}
	t.Skip("skipping: no safe adjacent cloud_security_posture version pair available in the Fleet registry for in-place version update test")
	return "", ""
}

func isSafeAdjacentCSPMUpgrade(from, to string) bool {
	fromVer, err1 := semver.NewVersion(from)
	toVer, err2 := semver.NewVersion(to)
	if err1 != nil || err2 != nil || !toVer.GreaterThan(fromVer) {
		return false
	}
	if fromVer.Major() != toVer.Major() {
		return false
	}
	if toVer.Minor()-fromVer.Minor() == 1 && fromVer.Patch() == 0 && toVer.Patch() == 0 {
		return true
	}
	if fromVer.Minor() == toVer.Minor() && toVer.Patch()-fromVer.Patch() == 1 {
		return true
	}
	return false
}

func cspmPackageVersionKnown(ctx context.Context, fc *fleetclient.Client, version string) bool {
	pkg, diags := fleetclient.GetPackage(ctx, fc, cspmPackageName, version, managedIntegrationDefaultSpace)
	return !diags.HasError() && pkg != nil
}

func ensureCSPMPackageInstalledOptional(ctx context.Context, fc *fleetclient.Client, version string) error {
	if cspmPackageInstalledInDefaultSpace(ctx, fc, version) {
		return nil
	}
	return tryInstallCSPMPackage(ctx, fc, version)
}
