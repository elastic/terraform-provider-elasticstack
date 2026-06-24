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

package osquerysavedquery

import (
	"context"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func (m *osquerySavedQueryModel) toAPICreateRequest(ctx context.Context) (kbapi.OsqueryCreateSavedQueryJSONRequestBody, diag.Diagnostics) {
	savedQueryID := m.SavedQueryID.ValueString()
	query := m.Query.ValueString()
	interval := strconv.FormatInt(m.Interval.ValueInt64(), 10)
	body := kbapi.OsqueryCreateSavedQueryJSONRequestBody{
		Id:       &savedQueryID,
		Query:    &query,
		Interval: &interval,
	}

	optional, diags := m.managedOptionalAPIFields(ctx)
	if diags.HasError() {
		return body, diags
	}

	body.Description = optional.description
	body.Platform = optional.platform
	body.Version = optional.version
	body.Snapshot = optional.snapshot
	body.Removed = optional.removed
	body.EcsMapping = optional.ecsMapping

	return body, diags
}

func (m *osquerySavedQueryModel) toAPIUpdateRequest(ctx context.Context, prior *osquerySavedQueryModel) (kbapi.OsqueryUpdateSavedQueryJSONRequestBody, diag.Diagnostics) {
	savedQueryID := m.SavedQueryID.ValueString()
	query := m.Query.ValueString()
	interval := strconv.FormatInt(m.Interval.ValueInt64(), 10)
	body := kbapi.OsqueryUpdateSavedQueryJSONRequestBody{
		Id:       &savedQueryID,
		Query:    &query,
		Interval: &interval,
	}

	optional, diags := m.managedOptionalAPIFields(ctx)
	if diags.HasError() {
		return body, diags
	}

	body.Description = optional.description
	body.Platform = optional.platform
	body.Version = optional.version
	body.Snapshot = optional.snapshot
	body.Removed = optional.removed
	body.EcsMapping = optional.ecsMapping

	if prior != nil {
		m.applyRemovedOptionalFields(prior, &body)
	}

	return body, diags
}

func (m *osquerySavedQueryModel) applyRemovedOptionalFields(prior *osquerySavedQueryModel, body *kbapi.OsqueryUpdateSavedQueryJSONRequestBody) {
	if body.Description == nil && typeutils.IsKnown(prior.Description) {
		empty := kbapi.SecurityOsqueryAPISavedQueryDescription("")
		body.Description = &empty
	}

	if body.Platform == nil && typeutils.IsKnown(prior.Platform) {
		empty := kbapi.SecurityOsqueryAPIPlatform("")
		body.Platform = &empty
	}

	if body.Version == nil && typeutils.IsKnown(prior.Version) {
		empty := kbapi.SecurityOsqueryAPIVersion("")
		body.Version = &empty
	}

	if body.EcsMapping == nil && typeutils.IsKnown(prior.EcsMapping) && !prior.EcsMapping.IsNull() {
		empty := kbapi.SecurityOsqueryAPIECSMapping{}
		body.EcsMapping = &empty
	}
}

type managedOptionalAPIFields struct {
	description *kbapi.SecurityOsqueryAPISavedQueryDescription
	platform    *kbapi.SecurityOsqueryAPIPlatform
	version     *kbapi.SecurityOsqueryAPIVersion
	snapshot    *kbapi.SecurityOsqueryAPISnapshot
	removed     *kbapi.SecurityOsqueryAPIRemoved
	ecsMapping  *kbapi.SecurityOsqueryAPIECSMapping
}

func (m *osquerySavedQueryModel) managedOptionalAPIFields(ctx context.Context) (managedOptionalAPIFields, diag.Diagnostics) {
	var (
		result managedOptionalAPIFields
		diags  diag.Diagnostics
	)

	if typeutils.IsKnown(m.Description) {
		description := m.Description.ValueString()
		result.description = &description
	}

	platform, platformDiags := platformToAPI(ctx, m.Platform)
	diags.Append(platformDiags...)
	if diags.HasError() {
		return result, diags
	}
	result.platform = platform

	if typeutils.IsKnown(m.Version) {
		version := m.Version.ValueString()
		result.version = &version
	}

	if typeutils.IsKnown(m.Snapshot) {
		snapshot := m.Snapshot.ValueBool()
		result.snapshot = &snapshot
	}

	if typeutils.IsKnown(m.Removed) {
		removed := m.Removed.ValueBool()
		result.removed = &removed
	}

	if typeutils.IsKnown(m.EcsMapping) && !m.EcsMapping.IsNull() {
		ecsMapping, ecsDiags := ecsMappingToAPI(ctx, m.EcsMapping)
		diags.Append(ecsDiags...)
		if diags.HasError() {
			return result, diags
		}
		result.ecsMapping = ecsMapping
	}

	return result, diags
}

func ecsMappingToAPI(ctx context.Context, mapping types.Map) (*kbapi.SecurityOsqueryAPIECSMapping, diag.Diagnostics) {
	var diags diag.Diagnostics

	elems := mapping.Elements()
	if len(elems) == 0 {
		empty := kbapi.SecurityOsqueryAPIECSMapping{}
		return &empty, diags
	}

	result := make(kbapi.SecurityOsqueryAPIECSMapping, len(elems))
	for key, value := range elems {
		obj, ok := value.(types.Object)
		if !ok {
			diags.AddError("Invalid ecs_mapping element", "Expected object value for ecs_mapping map entry.")
			return nil, diags
		}

		var elem ecsMapping
		diags.Append(obj.As(ctx, &elem, basetypes.ObjectAsOptions{})...)
		if diags.HasError() {
			return nil, diags
		}

		item, itemDiags := elem.toAPIType()
		diags.Append(itemDiags...)
		if diags.HasError() {
			return nil, diags
		}
		result[key] = item
	}

	return &result, diags
}
