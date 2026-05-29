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
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MinSupportedVersion is the minimum Elasticsearch version supported by the
// connector sync job create action. Although the connector resource itself
// (POST/PUT/GET /_connector) is GA from 8.12.0, the POST /_connector/_sync_job
// body validation on 8.12.x–8.15.x rejects the on-wire field name produced by
// the current typed Elasticsearch Go client; the API stabilized on the field
// shape we send starting with 8.16.0.
var MinSupportedVersion = version.Must(version.NewVersion("8.16.0"))

// Model holds the Terraform configuration for the connector sync job create action.
// The elasticsearch_connection and timeouts blocks are provided by the embedded
// envelope fields and injected into the schema by [entitycore.NewElasticsearchAction].
type Model struct {
	entitycore.ElasticsearchConnectionField
	entitycore.ActionTimeoutsField

	ConnectorID       types.String `tfsdk:"connector_id"`
	JobType           types.String `tfsdk:"job_type"`
	TriggerMethod     types.String `tfsdk:"trigger_method"`
	WaitForCompletion types.Bool   `tfsdk:"wait_for_completion"`
}

var _ entitycore.WithVersionRequirements = Model{}

// GetVersionRequirements satisfies [entitycore.WithVersionRequirements].
func (Model) GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics) {
	return []entitycore.VersionRequirement{{
		MinVersion:   *MinSupportedVersion,
		ErrorMessage: "elasticstack_elasticsearch_connector_sync_job_create requires Elasticsearch 8.16.0 or later (connector sync job APIs).",
	}}, nil
}
