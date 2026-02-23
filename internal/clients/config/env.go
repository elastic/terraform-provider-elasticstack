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

package config

import (
	"net/http"

	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
)

func NewFromEnv(version string) Client {
	ua := buildUserAgent(version)
	base := baseConfig{
		UserAgent: ua,
		Header:    http.Header{"User-Agent": []string{ua}},
	}.withEnvironmentOverrides()

	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg := base.toElasticsearchConfig().withEnvironmentOverrides()
	client.Elasticsearch = schemautil.Pointer(esCfg.toElasticsearchConfiguration())

	kibanaCfg := base.toKibanaConfig().withEnvironmentOverrides()
	client.Kibana = (*kibana.Config)(&kibanaCfg)

	kibanaOapiCfg := base.toKibanaOapiConfig().withEnvironmentOverrides()
	client.KibanaOapi = (*kibanaoapi.Config)(&kibanaOapiCfg)

	fleetCfg := kibanaOapiCfg.toFleetConfig().withEnvironmentOverrides()
	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client
}
