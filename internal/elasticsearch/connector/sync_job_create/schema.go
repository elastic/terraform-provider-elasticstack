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

package sync_job_create

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const schemaMarkdownDescription = `Creates an Elasticsearch connector sync job on demand. **Requires Terraform 1.14+** (provider-defined actions).

Invokes ` + "`POST /_connector/_sync_job`" + ` for an existing connector. When ` + "`wait_for_completion`" + ` is ` + "`true`" + `, polls ` +
	"`GET /_connector/_sync_job/{id}`" + ` until the job reaches a terminal status or the invoke timeout elapses. ` +
	`Sync job documents are retained after the action completes.`

// GetSchema returns the action schema for connector sync job create. The
// elasticsearch_connection and timeouts blocks are added by
// [entitycore.NewElasticsearchAction] and MUST NOT be declared here.
func GetSchema(_ context.Context) actionschema.Schema {
	return actionschema.Schema{
		MarkdownDescription: schemaMarkdownDescription,
		Attributes: map[string]actionschema.Attribute{
			"connector_id": actionschema.StringAttribute{
				MarkdownDescription: "The id of the connector to sync.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"job_type": actionschema.StringAttribute{
				MarkdownDescription: "Sync job type: `full`, `incremental`, or `access_control`. Defaults to `full` when omitted.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("full", "incremental", "access_control"),
				},
			},
			"trigger_method": actionschema.StringAttribute{
				MarkdownDescription: "How the sync job was triggered: `on_demand` or `scheduled`. Defaults to `on_demand` when omitted.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("on_demand", "scheduled"),
				},
			},
			"wait_for_completion": actionschema.BoolAttribute{
				MarkdownDescription: "When `true`, blocks until the sync job reaches a terminal status (`completed`, `cancelled`, `error`, or `suspended`). Defaults to `false` when omitted.",
				Optional:            true,
			},
		},
	}
}
