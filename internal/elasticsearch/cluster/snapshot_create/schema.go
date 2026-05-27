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

package snapshot_create

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/cluster"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const schemaMarkdownDescription = `Creates an Elasticsearch snapshot on demand. **Requires Terraform 1.14+** (provider-defined actions).

Invokes ` + "`POST /_snapshot/{repository}/{snapshot}`" + `. See the [create snapshot API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-snapshot-create).`

// GetSchema returns the action schema for snapshot create. The
// elasticsearch_connection and timeouts blocks are added by
// [entitycore.NewElasticsearchAction] and MUST NOT be declared here.
func GetSchema(_ context.Context) actionschema.Schema {
	return actionschema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Attributes: map[string]actionschema.Attribute{
			"repository": actionschema.StringAttribute{
				MarkdownDescription: "Name of the snapshot repository.",
				Required:            true,
			},
			"snapshot": actionschema.StringAttribute{
				MarkdownDescription: "Name to assign to the snapshot.",
				Required:            true,
			},
			"indices": actionschema.ListAttribute{
				MarkdownDescription: "Index patterns to include in the snapshot. All indices are included when omitted.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"include_global_state": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to include cluster state. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"ignore_unavailable": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to ignore missing or closed indices. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"partial": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to allow a partial snapshot when some shards are unavailable. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"feature_states": actionschema.ListAttribute{
				MarkdownDescription: "Feature states to include in the snapshot.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"expand_wildcards": actionschema.StringAttribute{
				MarkdownDescription: "Wildcard expansion for `indices`: `open`, `closed`, `hidden`, `none`, or `all`. Elasticsearch defaults to `open` when omitted.",
				Optional:            true,
				Validators: []validator.String{
					cluster.ExpandWildcardsValidator{},
				},
			},
			"metadata": actionschema.StringAttribute{
				MarkdownDescription: "JSON-encoded metadata attached to the snapshot.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"wait_for_completion": actionschema.BoolAttribute{
				MarkdownDescription: "When `true`, waits until snapshot creation completes. Defaults to `true` when omitted.",
				Optional:            true,
			},
		},
	}
}
