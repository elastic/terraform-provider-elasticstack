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

package snapshot_restore

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/schema"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	actiontimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const schemaMarkdownDescription = `Restores an Elasticsearch snapshot. **Requires Terraform 1.14+** (provider-defined actions).

Invokes ` + "`POST /_snapshot/{repository}/{snapshot}/_restore`" + `. See the [restore snapshot API documentation](https://www.elastic.co/docs/api/doc/elasticsearch/operation/operation-snapshot-restore).`

// GetSchema returns the action schema for snapshot restore.
func GetSchema(ctx context.Context) actionschema.Schema {
	return actionschema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Attributes: map[string]actionschema.Attribute{
			"repository": actionschema.StringAttribute{
				MarkdownDescription: "Name of the snapshot repository.",
				Required:            true,
			},
			"snapshot": actionschema.StringAttribute{
				MarkdownDescription: "Name of the snapshot to restore.",
				Required:            true,
			},
			"indices": actionschema.ListAttribute{
				MarkdownDescription: "Index patterns to restore. All indices in the snapshot are restored when omitted.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"include_global_state": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to restore cluster state. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"ignore_unavailable": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to ignore missing or closed indices. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"include_aliases": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to restore index aliases. Elasticsearch defaults to `true` when omitted.",
				Optional:            true,
			},
			"partial": actionschema.BoolAttribute{
				MarkdownDescription: "Whether to allow a partial restore when some shards are unavailable. Elasticsearch defaults to `false` when omitted.",
				Optional:            true,
			},
			"feature_states": actionschema.ListAttribute{
				MarkdownDescription: "Feature states to restore.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"rename_pattern": actionschema.StringAttribute{
				MarkdownDescription: "Regular expression pattern used to rename restored indices.",
				Optional:            true,
			},
			"rename_replacement": actionschema.StringAttribute{
				MarkdownDescription: "Replacement string applied with `rename_pattern`.",
				Optional:            true,
			},
			"ignore_index_settings": actionschema.ListAttribute{
				MarkdownDescription: "Index settings to ignore during restore.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"index_settings": actionschema.StringAttribute{
				MarkdownDescription: "JSON-encoded index settings overrides applied during restore.",
				Optional:            true,
				CustomType:          jsontypes.NormalizedType{},
			},
			"wait_for_completion": actionschema.BoolAttribute{
				MarkdownDescription: "When `true`, waits until the restore completes. Defaults to `true` when omitted.",
				Optional:            true,
			},
		},
		Blocks: map[string]actionschema.Block{
			"timeouts": actiontimeouts.Block(ctx),
			"elasticsearch_connection": schema.GetEsActionConnectionBlock(),
		},
	}
}
