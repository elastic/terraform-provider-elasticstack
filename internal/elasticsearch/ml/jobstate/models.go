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

package jobstate

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MLJobStateData struct {
	ID                      types.String         `tfsdk:"id"`
	ElasticsearchConnection types.List           `tfsdk:"elasticsearch_connection"`
	JobID                   types.String         `tfsdk:"job_id"`
	State                   types.String         `tfsdk:"state"`
	Force                   types.Bool           `tfsdk:"force"`
	Timeout                 customtypes.Duration `tfsdk:"job_timeout"`
	Timeouts                timeouts.Value       `tfsdk:"timeouts"`
}

// MLJobStats represents the statistics structure for an ML job
type MLJobStats struct {
	Jobs []MLJob `json:"jobs"`
}

// MLJob represents a single ML job in the stats response
type MLJob struct {
	JobID string     `json:"job_id"`
	State string     `json:"state"`
	Node  *MLJobNode `json:"node,omitempty"`
}

// MLJobNode represents the node information for an ML job
type MLJobNode struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Attributes map[string]any `json:"attributes"`
}
