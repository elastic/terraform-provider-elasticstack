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

package panelkit

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TimeRangeToAPI serializes a panel-level TimeRangeModel into the API time range schema.
// Returns nil when tr is nil (or its required from/to are not known) so the field is omitted
// from the API payload — matching the "omit when practitioner unset" write semantics.
func TimeRangeToAPI(tr *models.TimeRangeModel) *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema {
	if tr == nil {
		return nil
	}
	if !typeutils.IsKnown(tr.From) || !typeutils.IsKnown(tr.To) {
		return nil
	}
	out := &kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema{
		From: tr.From.ValueString(),
		To:   tr.To.ValueString(),
	}
	if typeutils.IsKnown(tr.Mode) {
		m := kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchemaMode(tr.Mode.ValueString())
		out.Mode = &m
	}
	return out
}

// TimeRangeFromAPI maps the API time range into state with REQ-009 null-preservation.
//
// prior is the prior-state panel TimeRangeModel (may be nil). Returns nil when the practitioner
// omitted the panel time_range (prior == nil), even when the API echoes an inherited dashboard
// range — this prevents drift from Kibana server-side inheritance. When prior is non-nil, the
// API from/to are populated and mode is preserved: API mode wins, else prior mode if known,
// else null (mirrors the dashboard root time_range mode-preservation).
func TimeRangeFromAPI(prior *models.TimeRangeModel, api *kbapi.KibanaHTTPAPIsKbnEsQueryServerTimeRangeSchema) *models.TimeRangeModel {
	if prior == nil {
		return nil
	}
	if api == nil {
		return nil
	}
	tr := &models.TimeRangeModel{
		From: types.StringValue(api.From),
		To:   types.StringValue(api.To),
	}
	switch {
	case api.Mode != nil:
		tr.Mode = types.StringValue(string(*api.Mode))
	case typeutils.IsKnown(prior.Mode):
		tr.Mode = prior.Mode
	default:
		tr.Mode = types.StringNull()
	}
	return tr
}
