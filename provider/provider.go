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

package provider

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/security"
	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/transform"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana"
	providerSchema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const esKeyName = "elasticsearch"
const kbKeyName = "kibana"
const fleetKeyName = "fleet"

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			esKeyName:    providerSchema.GetEsConnectionSchema(esKeyName, true),
			kbKeyName:    providerSchema.GetKibanaConnectionSchema(),
			fleetKeyName: providerSchema.GetFleetConnectionSchema(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"elasticstack_elasticsearch_security_role": security.DataSourceRole(),
			"elasticstack_elasticsearch_security_user": security.DataSourceUser(),
			"elasticstack_elasticsearch_info":          cluster.DataSourceClusterInfo(),

			"elasticstack_kibana_action_connector": kibana.DataSourceConnector(),
			"elasticstack_kibana_security_role":    kibana.DataSourceRole(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"elasticstack_elasticsearch_cluster_settings":   cluster.ResourceSettings(),
			"elasticstack_elasticsearch_component_template": index.ResourceComponentTemplate(),

			"elasticstack_elasticsearch_snapshot_lifecycle":  cluster.ResourceSlm(),
			"elasticstack_elasticsearch_snapshot_repository": cluster.ResourceSnapshotRepository(),
			"elasticstack_elasticsearch_transform":           transform.ResourceTransform(),

			"elasticstack_kibana_space":         kibana.ResourceSpace(),
			"elasticstack_kibana_security_role": kibana.ResourceRole(),
		},
	}

	p.ConfigureContextFunc = clients.NewAPIClientFuncFromSDK(version)

	return p
}
