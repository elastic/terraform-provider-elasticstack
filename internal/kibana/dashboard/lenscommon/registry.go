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

package lenscommon

import (
	"sort"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

var (
	convertersByType map[string]VizConverter
	sortedConverters []VizConverter // cached sorted snapshot returned by All; rebuilt on (Un)Register
)

// UnregisterVizConverter removes the converter registered for vizType and returns it (nil if none).
// Intended for tests that need to simulate a missing chart implementation.
func UnregisterVizConverter(vizType string) VizConverter {
	if convertersByType == nil {
		return nil
	}
	c := convertersByType[vizType]
	delete(convertersByType, vizType)
	rebuildSortedConverters()
	return c
}

// Register adds a VizConverter keyed by VizType(). Later registration replaces an earlier one with the same VizType().
func Register(c VizConverter) {
	if convertersByType == nil {
		convertersByType = make(map[string]VizConverter, 16)
	}
	convertersByType[c.VizType()] = c
	rebuildSortedConverters()
}

func rebuildSortedConverters() {
	if len(convertersByType) == 0 {
		sortedConverters = nil
		return
	}
	keys := make([]string, 0, len(convertersByType))
	for k := range convertersByType {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]VizConverter, 0, len(keys))
	for _, k := range keys {
		out = append(out, convertersByType[k])
	}
	sortedConverters = out
}

// canonicalVizTypeForOpaqueAttrs maps shorthand or legacy strings sometimes found in opaque
// panel config_json attributes["type"] to the kbapi discriminator strings used as registry keys.
func canonicalVizTypeForOpaqueAttrs(vizType string) string {
	switch vizType {
	case "tagcloud":
		return "tag_cloud"
	case "datatable":
		return "data_table"
	default:
		return vizType
	}
}

// ForType returns the converter registered for vizType, or nil if none.
// vizType may be either the kbapi discriminator (for example "tag_cloud") or legacy opaque-json
// spellings still accepted by practitioners ("tagcloud").
func ForType(vizType string) VizConverter {
	if convertersByType == nil {
		return nil
	}
	if c := convertersByType[vizType]; c != nil {
		return c
	}
	if alt := canonicalVizTypeForOpaqueAttrs(vizType); alt != vizType {
		return convertersByType[alt]
	}
	return nil
}

// FirstForBlocks returns the first converter (in stable All() order) whose HandlesBlocks reports true.
func FirstForBlocks(blocks *models.LensByValueChartBlocks) (VizConverter, bool) {
	if blocks == nil {
		return nil, false
	}
	for _, c := range All() {
		if c.HandlesBlocks(blocks) {
			return c, true
		}
	}
	return nil, false
}

// All returns every registered converter sorted by VizType().
// The returned slice is a cached snapshot rebuilt on Register/Unregister; callers must not mutate it.
func All() []VizConverter {
	return sortedConverters
}
