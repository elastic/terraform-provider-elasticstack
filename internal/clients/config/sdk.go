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
	"github.com/disaster37/go-kibana-rest/v8"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/fleet"
	kibanaoapi "github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanaoapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	esKey           string = "elasticsearch"
	esConnectionKey string = "elasticsearch_connection"
)

func NewFromSDK(d *schema.ResourceData, version string) (Client, diag.Diagnostics) {
	return newFromSDK(d, version, esKey)
}

func NewFromSDKResource(d *schema.ResourceData, version string) (*Client, diag.Diagnostics) {
	if _, ok := d.GetOk(esConnectionKey); !ok {
		return nil, nil
	}

	client, diags := newFromSDK(d, version, esConnectionKey)
	return &client, diags
}

func newFromSDK(d *schema.ResourceData, version, esConfigKey string) (Client, diag.Diagnostics) {
	base := newBaseConfigFromSDK(d, version, esConfigKey)
	client := Client{
		UserAgent: base.UserAgent,
	}

	esCfg, diags := newElasticsearchConfigFromSDK(d, base, esConfigKey, true)
	if diags.HasError() {
		return Client{}, diags
	}

	if esCfg != nil {
		client.Elasticsearch = schemautil.Pointer(esCfg.toElasticsearchConfiguration())
	}

	kibanaCfg, diags := newKibanaConfigFromSDK(d, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Kibana = (*kibana.Config)(&kibanaCfg)

	kibanaOapiCfg, diags := newKibanaOapiConfigFromSDK(d, base)
	if diags.HasError() {
		return Client{}, diags
	}

	client.KibanaOapi = (*kibanaoapi.Config)(&kibanaOapiCfg)

	fleetCfg, diags := newFleetConfigFromSDK(d, kibanaOapiCfg)
	if diags.HasError() {
		return Client{}, diags
	}

	client.Fleet = (*fleet.Config)(&fleetCfg)

	return client, nil
}
