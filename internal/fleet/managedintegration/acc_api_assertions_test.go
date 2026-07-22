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
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	fleetclient "github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/fleet/policyshape"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	cspmMappedInputKey             = "cspm-cloudbeat/cis_aws"
	cspmFindingsStreamKey          = "cloud_security_posture.findings"
	managedIntegrationDefaultSpace = "default"
)

func managedIntegrationPolicyFromState(s *terraform.State, resourceName string) (policyID, spaceID string, err error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok || rs.Primary == nil {
		return "", "", fmt.Errorf("resource %s not found in state", resourceName)
	}
	policyID = rs.Primary.Attributes["policy_id"]
	spaceID = managedIntegrationDefaultSpace
	if id, diags := clients.CompositeIDFromStr(rs.Primary.ID); !diags.HasError() && id != nil {
		spaceID = id.ClusterID
	}
	return policyID, spaceID, nil
}

func readManagedIntegrationAPI(ctx context.Context, spaceID, policyID string) (*kbapi.KibanaHTTPAPIsManagedIntegration, error) {
	client, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		return nil, err
	}
	fc := client.GetFleetClient()
	item, diags := fleetclient.ReadManagedIntegration(ctx, fc, spaceID, policyID)
	if diags.HasError() {
		return nil, diagutil.FwDiagsAsError(diags)
	}
	if item == nil {
		return nil, fmt.Errorf("managed integration %s not found", policyID)
	}
	return item, nil
}

func testCheckManagedIntegrationConditionsPersisted(resourceName, inputCondition, streamCondition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		in, ok := item.Inputs[cspmMappedInputKey]
		if !ok {
			return fmt.Errorf("input %q missing from managed integration %s API response", cspmMappedInputKey, policyID)
		}
		if in.Condition == nil || *in.Condition != inputCondition {
			got := ""
			if in.Condition != nil {
				got = *in.Condition
			}
			return fmt.Errorf("managed integration %s: expected input condition %q, got %q", policyID, inputCondition, got)
		}
		if in.Streams == nil {
			return fmt.Errorf("managed integration %s: input %q has no streams in API response", policyID, cspmMappedInputKey)
		}
		stream, ok := (*in.Streams)[cspmFindingsStreamKey]
		if !ok {
			return fmt.Errorf("stream %q missing from managed integration %s API response", cspmFindingsStreamKey, policyID)
		}
		if stream.Condition == nil || *stream.Condition != streamCondition {
			got := ""
			if stream.Condition != nil {
				got = *stream.Condition
			}
			return fmt.Errorf("managed integration %s: expected stream condition %q, got %q", policyID, streamCondition, got)
		}
		return nil
	}
}

func testCheckManagedIntegrationGlobalDataTagsPersisted(resourceName string, stringTags map[string]string, numberTags map[string]float64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		if item.GlobalDataTags == nil {
			return fmt.Errorf("managed integration %s: global_data_tags missing from API response", policyID)
		}
		gotString := map[string]string{}
		gotNumber := map[string]float64{}
		for _, tag := range *item.GlobalDataTags {
			raw, err := json.Marshal(tag.Value)
			if err != nil {
				return fmt.Errorf("managed integration %s: marshal tag %q value: %w", policyID, tag.Name, err)
			}
			var asString string
			if err := json.Unmarshal(raw, &asString); err == nil {
				gotString[tag.Name] = asString
				continue
			}
			var asNumber float64
			if err := json.Unmarshal(raw, &asNumber); err == nil {
				gotNumber[tag.Name] = asNumber
				continue
			}
			return fmt.Errorf("managed integration %s: unexpected global_data_tags value shape for %q: %s", policyID, tag.Name, string(raw))
		}
		for k, want := range stringTags {
			if gotString[k] != want {
				return fmt.Errorf("managed integration %s: global_data_tags[%q] string: got %q, want %q", policyID, k, gotString[k], want)
			}
		}
		for k, want := range numberTags {
			if gotNumber[k] != want {
				return fmt.Errorf("managed integration %s: global_data_tags[%q] number: got %v, want %v", policyID, k, gotNumber[k], want)
			}
		}
		return nil
	}
}

func testCheckManagedIntegrationStreamVarString(resourceName, varKey, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		in, ok := item.Inputs[cspmMappedInputKey]
		if !ok || in.Streams == nil {
			return fmt.Errorf("managed integration %s: input %q unavailable in API response", policyID, cspmMappedInputKey)
		}
		stream, ok := (*in.Streams)[cspmFindingsStreamKey]
		if !ok || stream.Vars == nil {
			return fmt.Errorf("managed integration %s: stream vars unavailable in API response", policyID)
		}
		v, ok := (*stream.Vars)[varKey]
		if !ok {
			return fmt.Errorf("managed integration %s: var %q missing from API response", policyID, varKey)
		}
		got, ok := managedIntegrationStreamVarString(v)
		if !ok {
			return fmt.Errorf("managed integration %s: var %q is not a decodable string value", policyID, varKey)
		}
		if got != want {
			return fmt.Errorf("managed integration %s: var %q: got %q, want %q", policyID, varKey, got, want)
		}
		return nil
	}
}

func testCheckManagedIntegrationNamePersisted(resourceName, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		if item.Name != expectedName {
			return fmt.Errorf("managed integration %s: expected name %q, got %q", policyID, expectedName, item.Name)
		}
		return nil
	}
}

func testCheckManagedIntegrationPackageVersionPersisted(resourceName, expectedVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		if item.Package.Version != expectedVersion {
			return fmt.Errorf("managed integration %s: expected package.version %q, got %q", policyID, expectedVersion, item.Package.Version)
		}
		return nil
	}
}

func testCheckManagedIntegrationUpdateExtrasPersisted(
	resourceName string,
	wantVars map[string]any,
	wantVarGroups map[string]string,
	wantPerms []string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(wantVars) == 0 && len(wantVarGroups) == 0 && len(wantPerms) == 0 {
			return fmt.Errorf("testCheckManagedIntegrationUpdateExtrasPersisted: at least one expected section must be non-empty")
		}
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		if len(wantVars) > 0 {
			if item.Vars == nil {
				return fmt.Errorf("managed integration %s: vars missing from API response", policyID)
			}
			got := policyshape.VarsAnyToMap(item.Vars)
			for k, want := range wantVars {
				gotVal, ok := got[k]
				if !ok {
					return fmt.Errorf("managed integration %s: vars[%q] missing from API; got keys %v", policyID, k, mapKeys(got))
				}
				if !managedIntegrationAPIValuesEqual(gotVal, want) {
					return fmt.Errorf("managed integration %s: vars[%q] want %v (%T), got %v (%T)", policyID, k, want, want, gotVal, gotVal)
				}
			}
		}
		if len(wantVarGroups) > 0 {
			if item.VarGroupSelections == nil {
				return fmt.Errorf("managed integration %s: var_group_selections missing from API response", policyID)
			}
			for k, want := range wantVarGroups {
				got, ok := (*item.VarGroupSelections)[k]
				if !ok || got != want {
					return fmt.Errorf("managed integration %s: var_group_selections[%q] want %q, got %q", policyID, k, want, got)
				}
			}
		}
		if len(wantPerms) > 0 {
			if item.AdditionalDatastreamsPermissions == nil {
				return fmt.Errorf("managed integration %s: additional_datastreams_permissions missing from API response", policyID)
			}
			got := *item.AdditionalDatastreamsPermissions
			if len(got) != len(wantPerms) {
				return fmt.Errorf("managed integration %s: additional_datastreams_permissions len want %d, got %d (%v)",
					policyID, len(wantPerms), len(got), got)
			}
			for i, want := range wantPerms {
				if got[i] != want {
					return fmt.Errorf("managed integration %s: additional_datastreams_permissions[%d] want %q, got %q",
						policyID, i, want, got[i])
				}
			}
		}
		return nil
	}
}

func managedIntegrationAPIValuesEqual(got, want any) bool {
	gotBytes, err1 := json.Marshal(got)
	wantBytes, err2 := json.Marshal(want)
	if err1 != nil || err2 != nil {
		return fmt.Sprint(got) == fmt.Sprint(want)
	}
	return string(gotBytes) == string(wantBytes)
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func testCheckCloudConnectorPersisted(resourceName, expectedConnectorID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		if item.CloudConnector == nil {
			return fmt.Errorf("managed integration %s: cloud_connector missing from API response", policyID)
		}
		if !item.CloudConnector.Enabled {
			return fmt.Errorf("managed integration %s: expected cloud_connector.enabled true, got false", policyID)
		}
		if item.CloudConnector.CloudConnectorId != expectedConnectorID {
			return fmt.Errorf("managed integration %s: expected cloud_connector_id %q, got %q",
				policyID, expectedConnectorID, item.CloudConnector.CloudConnectorId)
		}
		return nil
	}
}

// testCheckManagedIntegrationExternalIDStoredAsSecretRefOnAPI asserts the Fleet API
// persists aws.credentials.external_id as a secret reference while Terraform state
// keeps the configured plaintext (see acc_managed_integration_union_decode_test.go).
func testCheckManagedIntegrationExternalIDStoredAsSecretRefOnAPI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policyID, spaceID, err := managedIntegrationPolicyFromState(s, resourceName)
		if err != nil {
			return err
		}
		item, err := readManagedIntegrationAPI(context.Background(), spaceID, policyID)
		if err != nil {
			return err
		}
		v, err := managedIntegrationCSPMFindingsStreamVar(item, externalIDStreamVarKey)
		if err != nil {
			return fmt.Errorf("managed integration %s: %w", policyID, err)
		}
		refID, ok := managedIntegrationStreamVarSecretRefID(v)
		if !ok || refID == "" {
			return fmt.Errorf(
				"managed integration %s: expected %q stored as Fleet secret reference on API GET, but decode failed",
				policyID, externalIDStreamVarKey,
			)
		}
		return nil
	}
}
