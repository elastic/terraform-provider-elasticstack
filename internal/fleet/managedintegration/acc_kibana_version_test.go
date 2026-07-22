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
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/managedintegration"
	"github.com/hashicorp/go-version"
)

// kibanaBelowMinSkipReason formats the skip message when Kibana does not satisfy
// managedintegration.MinVersion via clients.KibanaScopedClient.EnforceMinVersion.
func kibanaBelowMinSkipReason(min *version.Version) string {
	return fmt.Sprintf("skipping: Kibana version is below the managed integration minimum %s", min.String())
}

// kibanaMeetsManagedIntegrationMinVersion evaluates the same Kibana /api/status
// version gate the resource uses in production (EnforceMinVersion), not the
// acceptance Elasticsearch cluster version.
func kibanaMeetsManagedIntegrationMinVersion(ctx context.Context, client *clients.KibanaScopedClient) (bool, error) {
	ok, diags := client.EnforceMinVersion(ctx, managedintegration.MinVersion)
	if diags.HasError() {
		return false, diagutil.FwDiagsAsError(diags)
	}
	return ok, nil
}

func skipUnlessKibanaMeetsManagedIntegrationMinVersion(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		return
	}
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client for version gate: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ok, err := kibanaMeetsManagedIntegrationMinVersion(ctx, client)
	if err != nil {
		t.Fatalf("failed to evaluate Kibana version gate: %v", err)
	}
	if !ok {
		t.Skip(kibanaBelowMinSkipReason(managedintegration.MinVersion))
	}
}

func skipIfKibanaMeetsManagedIntegrationMinVersion(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		return
	}
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("failed to create Kibana client for version gate: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ok, err := kibanaMeetsManagedIntegrationMinVersion(ctx, client)
	if err != nil {
		t.Fatalf("failed to evaluate Kibana version gate: %v", err)
	}
	if ok {
		t.Skipf("skipping: Kibana already meets managed integration minimum %s; version gate test requires Kibana below that floor",
			managedintegration.MinVersion.String())
	}
}

// skipUnlessManagedIntegrationLiveStack applies the preconditions for positive
// acceptance tests: Kibana >= MinVersion and a usable pinned CSPM package version.
// Call skipUnlessConfirmedCloud separately — topology gating stays independent.
func skipUnlessManagedIntegrationLiveStack(t *testing.T) {
	t.Helper()
	skipUnlessKibanaMeetsManagedIntegrationMinVersion(t)
	skipUnlessCSPMPinnedPackageAvailable(t)
}
