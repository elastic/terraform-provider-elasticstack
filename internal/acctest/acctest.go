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
	"fmt"
	"log"
	"os"
	"path"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/config"
)

var Providers map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	providerServerFactory, err := provider.ProtoV6ProviderServerFactory(context.Background(), provider.AccTestVersion)
	if err != nil {
		log.Fatal(err)
	}
	Providers = map[string]func() (tfprotov6.ProviderServer, error){
		"elasticstack": func() (tfprotov6.ProviderServer, error) {
			server := providerServerFactory()
			if server == nil {
				return nil, fmt.Errorf("provider server factory returned nil")
			}
			return server, nil
		},
	}
}

func PreCheck(t *testing.T) {
	_, elasticsearchEndpointsOk := os.LookupEnv("ELASTICSEARCH_ENDPOINTS")
	_, kibanaEndpointOk := os.LookupEnv("KIBANA_ENDPOINT")
	_, userOk := os.LookupEnv("ELASTICSEARCH_USERNAME")
	_, passOk := os.LookupEnv("ELASTICSEARCH_PASSWORD")
	_, apiKeyOk := os.LookupEnv("ELASTICSEARCH_API_KEY")
	_, kbUserOk := os.LookupEnv("KIBANA_USERNAME")
	_, kbPassOk := os.LookupEnv("KIBANA_PASSWORD")
	_, kbAPIKeyOk := os.LookupEnv("KIBANA_API_KEY")

	if !elasticsearchEndpointsOk {
		t.Fatal("ELASTICSEARCH_ENDPOINTS must be set for acceptance tests to run")
	}

	if !kibanaEndpointOk {
		t.Fatal("KIBANA_ENDPOINT must be set for acceptance tests to run")
	}

	authOk := (userOk && passOk) || (kbUserOk && kbPassOk) || apiKeyOk || kbAPIKeyOk
	if !authOk {
		t.Fatal("ELASTICSEARCH_USERNAME and ELASTICSEARCH_PASSWORD, or KIBANA_USERNAME and KIBANA_PASSWORD, or ELASTICSEARCH_API_KEY, or KIBANA_API_KEY must be set for acceptance tests to run")
	}
}

func NamedTestCaseDirectory(name string) config.TestStepConfigFunc {
	return func(tscr config.TestStepConfigRequest) string {
		return path.Join(config.TestNameDirectory()(tscr), name)
	}
}
