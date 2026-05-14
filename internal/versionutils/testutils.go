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

package versionutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/hashicorp/go-version"
)

// Flavor selects deployment flavor requirements for acceptance-test skips.
type Flavor int

const (
	FlavorAny Flavor = iota
	FlavorStateful
	FlavorServerless
)

func (f Flavor) String() string {
	switch f {
	case FlavorAny:
		return "Any"
	case FlavorStateful:
		return "Stateful"
	case FlavorServerless:
		return "Serverless"
	default:
		return fmt.Sprintf("Flavor(%d)", f)
	}
}

type serverInfoGetter func(context.Context) (*version.Version, string, error)

func fetchAcceptanceServerInfo(ctx context.Context) (*version.Version, string, error) {
	client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
	if err != nil {
		return nil, "", err
	}
	serverVersion, diags := client.ServerVersion(ctx)
	if diags.HasError() {
		return nil, "", fmt.Errorf("failed to parse elasticsearch server version: %v", diags)
	}
	buildFlavor, diags := client.ServerFlavor(ctx)
	if diags.HasError() {
		return nil, "", fmt.Errorf("failed to get elasticsearch server flavor: %v", diags)
	}
	return serverVersion, buildFlavor, nil
}

// checkSkip resolves server version and build flavor once via getServerInfo, then applies
// version and flavor rules. When both minVersion and constraints are empty and wantFlavor
// is FlavorAny, no client call is made.
func checkSkip(ctx context.Context, minVersion *version.Version, constraints version.Constraints, wantFlavor Flavor, getServerInfo serverInfoGetter) (skip bool, skipReason string, err error) {
	needVersion := minVersion != nil || len(constraints) > 0
	needFlavor := wantFlavor != FlavorAny
	if !needVersion && !needFlavor {
		return false, "", nil
	}

	serverVer, buildFlavor, err := getServerInfo(ctx)
	if err != nil {
		return false, "", err
	}

	isServerless := buildFlavor == clients.ServerlessFlavor

	if !isServerless {
		if minVersion != nil && serverVer.LessThan(minVersion) {
			return true, fmt.Sprintf("elasticsearch version %s is below required minimum %s", serverVer.String(), minVersion.String()), nil
		}
		if len(constraints) > 0 && !constraints.Check(serverVer) {
			return true, fmt.Sprintf("elasticsearch version %s does not satisfy constraints %s", serverVer.String(), constraints.String()), nil
		}
	}

	switch wantFlavor {
	case FlavorAny:
		return false, "", nil
	case FlavorStateful:
		if isServerless {
			return true, "test requires stateful elasticsearch but cluster is serverless", nil
		}
	case FlavorServerless:
		if !isServerless {
			return true, "test requires serverless elasticsearch but cluster is stateful", nil
		}
	default:
		return false, "", fmt.Errorf("unknown acceptance test flavor %v (%s)", wantFlavor, wantFlavor.String())
	}

	return false, "", nil
}

func skipContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// SkipIfUnsupported skips the test when the acceptance Elasticsearch connection reports a
// version strictly below minVersion or a deployment flavor incompatible with flavor.
// Serverless clusters bypass minimum-version checks. Infrastructure failures call t.Fatal.
func SkipIfUnsupported(t *testing.T, minVersion *version.Version, flavor Flavor) {
	t.Helper()
	skip, reason, err := checkSkip(skipContext(t), minVersion, nil, flavor, fetchAcceptanceServerInfo)
	if err != nil {
		t.Fatal(err)
	}
	if skip {
		t.Skip(reason)
	}
}

// SkipIfUnsupportedConstraints skips the test when the acceptance Elasticsearch version does
// not satisfy constraints or the deployment flavor is incompatible. Serverless clusters bypass
// constraint checks. Infrastructure failures call t.Fatal.
func SkipIfUnsupportedConstraints(t *testing.T, constraints version.Constraints, flavor Flavor) {
	t.Helper()
	skip, reason, err := checkSkip(skipContext(t), nil, constraints, flavor, fetchAcceptanceServerInfo)
	if err != nil {
		t.Fatal(err)
	}
	if skip {
		t.Skip(reason)
	}
}

func CheckIfVersionIsUnsupported(minSupportedVersion *version.Version) func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return false, err
		}
		serverVersion, diags := client.ServerVersion(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
		}

		return serverVersion.LessThan(minSupportedVersion), nil
	}
}

func CheckIfVersionMeetsConstraints(constraints version.Constraints) func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return false, err
		}
		serverVersion, diags := client.ServerVersion(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to parse the elasticsearch version %v", diags)
		}

		return !constraints.Check(serverVersion), nil
	}
}

func CheckIfNotServerless() func() (bool, error) {
	return func() (b bool, err error) {
		client, err := clients.NewAcceptanceTestingElasticsearchScopedClient()
		if err != nil {
			return false, err
		}
		serverFlavor, diags := client.ServerFlavor(context.Background())
		if diags.HasError() {
			return false, fmt.Errorf("failed to get the elasticsearch flavor %v", diags)
		}

		return serverFlavor != clients.ServerlessFlavor, nil
	}
}
