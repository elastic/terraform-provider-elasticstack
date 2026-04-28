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
	"maps"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type kibanaConnectionAuth struct {
	apiKey   string
	username string
	password string
}

func acceptanceTestKibanaConnectionAuth() kibanaConnectionAuth {
	apiKey := os.Getenv("KIBANA_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ELASTICSEARCH_API_KEY")
	}

	username := os.Getenv("KIBANA_USERNAME")
	if username == "" {
		username = os.Getenv("ELASTICSEARCH_USERNAME")
	}

	password := os.Getenv("KIBANA_PASSWORD")
	if password == "" {
		password = os.Getenv("ELASTICSEARCH_PASSWORD")
	}

	return kibanaConnectionAuth{
		apiKey:   apiKey,
		username: username,
		password: password,
	}
}

func acceptanceTestKibanaEndpoint() string {
	return strings.TrimSpace(os.Getenv("KIBANA_ENDPOINT"))
}

func KibanaConnectionVariables(additional ...config.Variables) config.Variables {
	auth := acceptanceTestKibanaConnectionAuth()
	vars := config.Variables{
		"kibana_endpoints": config.ListVariable(config.StringVariable(acceptanceTestKibanaEndpoint())),
		"api_key":          config.StringVariable(auth.apiKey),
		"username":         config.StringVariable(auth.username),
		"password":         config.StringVariable(auth.password),
	}

	for _, extra := range additional {
		maps.Copy(vars, extra)
	}

	return vars
}

func KibanaConnectionAuthChecks(resourceName string) []resource.TestCheckFunc {
	auth := acceptanceTestKibanaConnectionAuth()
	if auth.apiKey != "" {
		return []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.api_key", auth.apiKey),
		}
	}

	return []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.username", auth.username),
		resource.TestCheckResourceAttr(resourceName, "kibana_connection.0.password", auth.password),
	}
}
