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

package cloudconnector_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
)

const (
	defaultFleetServerPolicyName = "fleet-server"
	defaultSpaceID               = "default"
)

// attachPackagePolicyToCloudConnector sets cloud_connector_id on the default
// Fleet Server package policy so delete-without-force is rejected by the API.
func attachPackagePolicyToCloudConnector(t *testing.T, connectorID string) {
	t.Helper()
	acctest.SkipIfNotAcceptanceTest(t)

	if connectorID == "" {
		t.Fatal("connectorID is empty")
	}

	ctx := context.Background()
	kbClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("creating kibana client: %v", err)
	}

	fleetClient := kbClient.GetFleetClient()
	existing, policyID, skipReason := findFleetServerPackagePolicy(ctx, fleetClient, defaultSpaceID)
	if skipReason != "" {
		t.Skip(skipReason)
	}

	updateReq, buildErr := packagePolicyUpdateRequest(existing, &connectorID)
	if buildErr != nil {
		t.Fatalf("building package policy update request: %v", buildErr)
	}

	if _, diags := fleetclient.UpdateDefendPackagePolicy(ctx, fleetClient, policyID, defaultSpaceID, updateReq); diags.HasError() {
		t.Fatalf("attaching cloud connector to package policy %q: %v", policyID, diagutil.FwDiagsAsError(diags))
	}

	t.Cleanup(func() {
		clearReq, clearErr := packagePolicyUpdateRequest(existing, nil)
		if clearErr != nil {
			return
		}
		_, _ = fleetclient.UpdateDefendPackagePolicy(ctx, fleetClient, policyID, defaultSpaceID, clearReq)
	})
}

func findFleetServerPackagePolicy(ctx context.Context, client *fleetclient.Client, spaceID string) (*kbapi.PackagePolicy, string, string) {
	perPage := float32(100)
	page := float32(1)
	resp, err := client.API.GetFleetPackagePoliciesWithResponse(ctx, &kbapi.GetFleetPackagePoliciesParams{
		PerPage: &perPage,
		Page:    &page,
	}, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, "", fmt.Sprintf("unable to list package policies for force_delete fixture: %v", err)
	}
	if resp.JSON200 == nil {
		return nil, "", "unable to list package policies for force_delete fixture: empty response"
	}

	for _, item := range resp.JSON200.Items {
		if item.Name != defaultFleetServerPolicyName {
			continue
		}
		policy, diags := fleetclient.GetDefendPackagePolicy(ctx, client, item.Id, spaceID)
		if diags.HasError() || policy == nil {
			return nil, "", fmt.Sprintf("unable to load package policy %q for force_delete fixture: %v", item.Id, diags)
		}
		return policy, item.Id, ""
	}

	return nil, "", fmt.Sprintf("no %q package policy found for force_delete fixture", defaultFleetServerPolicyName)
}

func packagePolicyUpdateRequest(existing *kbapi.PackagePolicy, cloudConnectorID *string) (kbapi.PackagePolicyRequestTypedInputs, error) {
	if existing == nil {
		return kbapi.PackagePolicyRequestTypedInputs{}, fmt.Errorf("existing package policy is nil")
	}
	if existing.Package == nil {
		return kbapi.PackagePolicyRequestTypedInputs{}, fmt.Errorf("existing package policy %q has no package metadata", existing.Id)
	}

	typedInputs, err := existing.Inputs.AsPackagePolicyTypedInputs()
	if err != nil {
		return kbapi.PackagePolicyRequestTypedInputs{}, fmt.Errorf("reading typed inputs: %w", err)
	}

	reqInputs := make([]kbapi.PackagePolicyRequestTypedInput, len(typedInputs))
	for i, input := range typedInputs {
		streams := make([]kbapi.PackagePolicyRequestTypedInputStream, len(input.Streams))
		for j, stream := range input.Streams {
			streams[j] = kbapi.PackagePolicyRequestTypedInputStream{
				Enabled:            stream.Enabled,
				Id:                 stream.Id,
				Vars:               stream.Vars,
				VarGroupSelections: stream.VarGroupSelections,
			}
		}
		reqInputs[i] = kbapi.PackagePolicyRequestTypedInput{
			Type:    input.Type,
			Enabled: input.Enabled,
			Id:      input.Id,
			Vars:    input.Vars,
			Streams: &streams,
		}
	}

	enabled := existing.Enabled
	return kbapi.PackagePolicyRequestTypedInputs{
		Name:             &existing.Name,
		Namespace:        existing.Namespace,
		Description:      existing.Description,
		Enabled:          &enabled,
		PolicyIds:        existing.PolicyIds,
		CloudConnectorId: cloudConnectorID,
		Package: &kbapi.PackagePolicyRequestPackage{
			Name:    existing.Package.Name,
			Version: existing.Package.Version,
		},
		Inputs: &reqInputs,
	}, nil
}
