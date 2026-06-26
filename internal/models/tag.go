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

package models

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

// Tag is the provider domain model for a Kibana tag returned by /api/tags.
type Tag struct {
	ID          string
	Name        string
	Color       string
	Description *string
	SpaceID     string
	CreatedAt   string
	UpdatedAt   string
	Managed     bool
}

// TagFromAPI maps kbapi tag attributes and as-code metadata into a Tag.
func TagFromAPI(id, spaceID string, data kbapi.KibanaHTTPAPIsKbnTagsAttributes, meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta) Tag {
	tag := Tag{
		ID:      id,
		SpaceID: spaceID,
		Name:    data.Name,
		Color:   data.Color,
	}
	if data.Description != nil {
		desc := *data.Description
		tag.Description = &desc
	}
	if meta.CreatedAt != nil {
		tag.CreatedAt = *meta.CreatedAt
	}
	if meta.UpdatedAt != nil {
		tag.UpdatedAt = *meta.UpdatedAt
	}
	if meta.Managed != nil {
		tag.Managed = *meta.Managed
	}
	return tag
}
