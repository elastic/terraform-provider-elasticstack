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

package acctest

import (
	"context"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// advancedSettingsMinVersion is the earliest Kibana version these tests run on.
// The behavior of the advanced settings routes and the Kibana error messages the
// tests assert were verified only on Kibana 9.5.6, so older stacks in the
// version matrix skip them.
var advancedSettingsMinVersion = version.Must(version.NewVersion("9.5.0"))

// AdvancedSettingsKibanaClient returns a Kibana client for the acceptance tests
// of Kibana advanced settings. The tests run only on stateful deployments at or
// above advancedSettingsMinVersion.
func AdvancedSettingsKibanaClient(t *testing.T) *kibanaoapi.Client {
	t.Helper()

	PreCheck(t)
	versionutils.SkipIfUnsupported(t, advancedSettingsMinVersion, versionutils.FlavorStateful)

	scopedClient, err := clients.NewAcceptanceTestingKibanaScopedClient()
	if err != nil {
		t.Fatalf("creating the acceptance testing Kibana client: %v", err)
	}
	return scopedClient.GetKibanaOapiClient()
}

// NewAdvancedSettingsTestSpace creates a Kibana space with a random ID and
// deletes it when the test ends. Advanced settings are shared stack state, so
// each test works in a space of its own.
func NewAdvancedSettingsTestSpace(t *testing.T, client *kibanaoapi.Client) string {
	t.Helper()

	spaceID := "advset-" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	if _, diags := kibanaoapi.CreateSpace(t.Context(), client, kbapi.PostSpacesSpaceJSONRequestBody{Id: spaceID, Name: spaceID}); diags.HasError() {
		t.Fatalf("creating space %q: %v", spaceID, diags)
	}
	t.Cleanup(func() {
		if diags := kibanaoapi.DeleteSpace(context.Background(), client, spaceID); diags.HasError() {
			t.Errorf("deleting space %q: %v", spaceID, diags)
		}
	})
	return spaceID
}
