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

package streams_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/acctest"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/versionutils"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/config"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var minVersionStreamsAcc = version.Must(version.NewVersion("9.2.0-SNAPSHOT"))

// prepareStreamsEnvironment runs before each test step. It calls the Kibana
// Streams resync API (to repair any inconsistent state from prior runs) and
// then attempts to configure the logs root stream with an explicit
// failure_store so child stream creation does not fail with
// "all ancestors have inherit configuration".
func prepareStreamsEnvironment(t *testing.T) {
	t.Helper()

	apiClient, err := clients.NewAcceptanceTestingClient()
	if err != nil {
		t.Logf("prepareStreamsEnvironment: could not create client: %v", err)
		return
	}
	kibanaClient, err := apiClient.GetKibanaOapiClient()
	if err != nil {
		t.Logf("prepareStreamsEnvironment: could not get Kibana client: %v", err)
		return
	}

	// Resync to repair any inconsistent Streams state.
	resp, err := kibanaClient.API.PostStreamsResyncWithResponse(context.Background(), kbapi.PostStreamsResyncJSONRequestBody{})
	if err != nil {
		t.Logf("prepareStreamsEnvironment: resync error: %v", err)
	} else {
		t.Logf("prepareStreamsEnvironment: resync status %d", resp.StatusCode())
	}

	// Configure the logs root stream with an explicit failure_store so that
	// child streams don't fail with "all ancestors have inherit configuration".
	// We first GET the current state of logs to preserve its existing wired
	// fields (which may include ES field aliases we must not remove).
	current, getDiags := kibanaoapi.GetStream(context.Background(), kibanaClient, "default", "logs")
	if getDiags.HasError() {
		t.Logf("prepareStreamsEnvironment: GET logs failed: %s", getDiags[0].Detail())
	} else if current != nil && current.Stream.Ingest != nil {
		t.Logf("prepareStreamsEnvironment: current logs failure_store=%s lifecycle=%s",
			string(current.Stream.Ingest.FailureStore),
			string(current.Stream.Ingest.Lifecycle),
		)
	}

	// Build the upsert, reusing whatever wired fields already exist to avoid
	// mapper_parsing_exception from field aliases.
	if current != nil {
		t.Logf("prepareStreamsEnvironment: logs stream exists (type=%s)", current.Stream.Type)
	} else {
		t.Logf("prepareStreamsEnvironment: logs stream not found — will create")
	}

	// As of kibana#251618 (9.4.0), fresh installs have logs.otel and logs.ecs
	// instead of a monolithic logs root. Wired child streams must live under
	// one of those roots. Ensure logs.otel has an explicit failure_store so
	// child stream creation doesn't fail with "all ancestors have inherit".
	logsRoot := "logs.otel"
	rootCurrent, _ := kibanaoapi.GetStream(context.Background(), kibanaClient, "default", logsRoot)
	if rootCurrent != nil {
		failureStore := "unknown"
		if rootCurrent.Stream.Ingest != nil {
			failureStore = string(rootCurrent.Stream.Ingest.FailureStore)
		}
		t.Logf("prepareStreamsEnvironment: %s exists (type=%s failure_store=%s)",
			logsRoot, rootCurrent.Stream.Type, failureStore)
	} else {
		t.Logf("prepareStreamsEnvironment: %s not found", logsRoot)
	}

	existingFields := json.RawMessage(`{}`)
	existingRouting := []kibanaoapi.StreamRoutingRule{}
	if rootCurrent != nil && rootCurrent.Stream.Ingest != nil && rootCurrent.Stream.Ingest.Wired != nil {
		if len(rootCurrent.Stream.Ingest.Wired.Fields) > 0 {
			existingFields = rootCurrent.Stream.Ingest.Wired.Fields
		}
		if len(rootCurrent.Stream.Ingest.Wired.Routing) > 0 {
			existingRouting = rootCurrent.Stream.Ingest.Wired.Routing
		}
	}

	req := kibanaoapi.StreamUpsertRequest{
		Stream: kibanaoapi.StreamDefinition{
			Type:        "wired",
			Description: "",
			Ingest: &kibanaoapi.StreamIngest{
				Processing:   kibanaoapi.StreamProcessing{Steps: json.RawMessage(`[]`)},
				Wired:        &kibanaoapi.StreamIngestWired{Fields: existingFields, Routing: existingRouting},
				Lifecycle:    json.RawMessage(`{"dsl":{"data_retention":"365d"}}`),
				FailureStore: json.RawMessage(`{"disabled":{}}`),
			},
		},
		Dashboards: []string{},
		Rules:      []string{},
		Queries:    []kibanaoapi.StreamQuery{},
	}
	_, diags := kibanaoapi.UpsertStream(context.Background(), kibanaClient, "default", logsRoot, req)
	if diags.HasError() {
		t.Logf("prepareStreamsEnvironment: configuring %s failed: %s — %s",
			logsRoot, diags[0].Summary(), diags[0].Detail())
	} else {
		t.Logf("prepareStreamsEnvironment: %s configured OK", logsRoot)
	}

	// Probe creating a test stream to surface any remaining issues.
	probeReq := kibanaoapi.StreamUpsertRequest{
		Stream: kibanaoapi.StreamDefinition{
			Type: "wired",
			Ingest: &kibanaoapi.StreamIngest{
				Processing:   kibanaoapi.StreamProcessing{Steps: json.RawMessage(`[]`)},
				Wired:        &kibanaoapi.StreamIngestWired{Fields: json.RawMessage(`{}`), Routing: []kibanaoapi.StreamRoutingRule{}},
				Lifecycle:    json.RawMessage(`{"dsl":{"data_retention":"365d"}}`),
				FailureStore: json.RawMessage(`{"disabled":{}}`),
			},
		},
		Dashboards: []string{},
		Rules:      []string{},
		Queries:    []kibanaoapi.StreamQuery{},
	}
	probeName := logsRoot + ".__tfacc_probe__"
	_, probeDiags := kibanaoapi.UpsertStream(context.Background(), kibanaClient, "default", probeName, probeReq)
	if probeDiags.HasError() {
		t.Logf("prepareStreamsEnvironment: probe stream creation failed: %s", probeDiags[0].Detail())
	} else {
		t.Logf("prepareStreamsEnvironment: probe stream created OK — environment ready")
		_ = kibanaoapi.DeleteStream(context.Background(), kibanaClient, "default", probeName)
	}
}

func TestAccResourceKibanaStreamWired(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	skipFn := versionutils.CheckIfVersionIsUnsupported(minVersionStreamsAcc)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			prepareStreamsEnvironment(t)
		},
		Steps: []resource.TestStep{
			// Step 1: create a minimal wired stream.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_stream.wired", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "name", "logs.otel.testacc"+suffix),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "space_id", "default"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "description", "Test wired stream"),
					// Optional attributes absent — verify empty/default state.
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.processing_steps.#", "0"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_stream.wired", "dashboards"),
					resource.TestCheckNoResourceAttr("elasticstack_kibana_stream.wired", "queries"),
				),
			},
			// Step 2: add a processing step — assert the step JSON value (not just count).
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "description", "Updated wired stream"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.processing_steps.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_stream.wired", "wired_config.processing_steps.0.json"),
				),
			},
			// Step 3: full update — lifecycle, failure_store, index settings, attached query.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_full"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "description", "Fully-configured wired stream"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.processing_steps.#", "1"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_stream.wired", "wired_config.lifecycle_json"),
					resource.TestCheckResourceAttrSet("elasticstack_kibana_stream.wired", "wired_config.failure_store_json"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.index_number_of_shards", "1"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.index_number_of_replicas", "0"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.wired", "wired_config.index_refresh_interval", "5s"),
				),
			},
			// Step 4: import from the fully-configured state.
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update_full"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_stream.wired",
			},
		},
	})
}

// checkQueryStreamsEnabled returns a SkipFunc that skips when Kibana query
// streams are not enabled (HTTP 422 "Streams are not enabled for Query streams").
func checkQueryStreamsEnabled() func() (bool, error) {
	return func() (bool, error) {
		apiClient, err := clients.NewAcceptanceTestingClient()
		if err != nil {
			return false, err
		}
		kibanaClient, err := apiClient.GetKibanaOapiClient()
		if err != nil {
			return false, err
		}
		probe := kibanaoapi.StreamUpsertRequest{
			Stream: kibanaoapi.StreamDefinition{
				Type:  "query",
				Query: &kibanaoapi.StreamQueryESQLDef{Esql: "FROM logs* | LIMIT 1"},
			},
			Dashboards: []string{},
			Rules:      []string{},
			Queries:    []kibanaoapi.StreamQuery{},
		}
		_, diags := kibanaoapi.UpsertStream(context.Background(), kibanaClient, "default", "logs.__tfacc_query_probe__", probe)
		if diags.HasError() {
			if strings.Contains(diags[0].Detail()+diags[0].Summary(), "not enabled") {
				return true, nil
			}
		}
		_ = kibanaoapi.DeleteStream(context.Background(), kibanaClient, "default", "logs.__tfacc_query_probe__")
		return false, nil
	}
}

func TestAccResourceKibanaStreamQuery(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(6, sdkacctest.CharSetAlphaNum)
	skipFn := func() (bool, error) {
		if skip, err := versionutils.CheckIfVersionIsUnsupported(minVersionStreamsAcc)(); skip || err != nil {
			return skip, err
		}
		return checkQueryStreamsEnabled()()
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			prepareStreamsEnvironment(t)
		},
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("elasticstack_kibana_stream.query", "id"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.query", "name", "logs.otel.testacc"+suffix+".view"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.query", "query_config.esql", "FROM logs* | LIMIT 10"),
					// view is Optional+Computed; when unset it should be stored as "".
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.query", "query_config.view", ""),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.query", "description", "Updated query stream"),
					resource.TestCheckResourceAttr("elasticstack_kibana_stream.query", "query_config.esql", "FROM logs* | WHERE @timestamp > NOW() - 1 HOUR | LIMIT 10"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.Providers,
				SkipFunc:                 skipFn,
				ConfigDirectory:          acctest.NamedTestCaseDirectory("update"),
				ConfigVariables: config.Variables{
					"suffix": config.StringVariable(suffix),
				},
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "elasticstack_kibana_stream.query",
			},
		},
	})
}

