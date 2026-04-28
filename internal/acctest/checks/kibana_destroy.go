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

package checks

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// KibanaResourceDestroyCheck creates a CheckDestroy function for Kibana acceptance
// tests. It handles the repeated scaffolding: creating a scoped client, obtaining
// the Kibana OpenAPI client, and iterating the Terraform state for resources of
// the given type. The lookup function receives the plain resource ID and should
// return (true, nil) when the resource still exists, (false, nil) when it is gone,
// or (false, err) on an unexpected API error.
func KibanaResourceDestroyCheck(
	resourceType string,
	lookup func(ctx context.Context, client *kibanaoapi.Client, id string) (bool, error),
) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		oapiClient, err := apiClient.GetKibanaOapiClient()
		if err != nil {
			return err
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			exists, err := lookup(context.Background(), oapiClient, rs.Primary.ID)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("%s (%s) still exists", resourceType, rs.Primary.ID)
			}
		}
		return nil
	}
}

// KibanaResourceDestroyCheckCompositeID creates a CheckDestroy function for
// Kibana resources whose state ID is a composite "spaceID/resourceID" string.
// It parses the composite ID and passes the space ID and resource ID separately
// to the lookup function, removing that repeated plumbing from each test file.
// The lookup function should return (true, nil) when the resource still exists,
// (false, nil) when it is gone, or (false, err) on an unexpected API error.
func KibanaResourceDestroyCheckCompositeID(
	resourceType string,
	lookup func(ctx context.Context, client *kibanaoapi.Client, spaceID, resourceID string) (bool, error),
) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
		if err != nil {
			return err
		}
		oapiClient, err := apiClient.GetKibanaOapiClient()
		if err != nil {
			return err
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
			compID, diags := clients.CompositeIDFromStr(rs.Primary.ID)
			if diags.HasError() {
				continue
			}
			exists, err := lookup(context.Background(), oapiClient, compID.ClusterID, compID.ResourceID)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("%s (%s) still exists", resourceType, rs.Primary.ID)
			}
		}
		return nil
	}
}
