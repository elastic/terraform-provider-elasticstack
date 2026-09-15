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

// PopulateXYMetricAxisDefault injects Kibana 9.6's omit-default axis value ("y")
// on XY y[] metric config_json when the practitioner omitted axis.
func PopulateXYMetricAxisDefault(model map[string]any) map[string]any {
	if model == nil {
		return model
	}
	if _, exists := model["axis"]; !exists {
		model["axis"] = "y"
	}
	return model
}

// PopulateXYMetricDefaults applies shared Lens metric defaults plus the XY-only
// axis omit-default. Used for XY y[] config_json, not datatable metrics[].
func PopulateXYMetricDefaults(model map[string]any) map[string]any {
	return PopulateXYMetricAxisDefault(PopulateLensMetricDefaults(model))
}

// PopulateDatatableMetricDefaults applies shared Lens metric defaults plus
// Kibana 9.6 datatable-only omit-defaults (visible, alignment). It does not
// inject axis — datatable metrics do not receive that key.
func PopulateDatatableMetricDefaults(model map[string]any) map[string]any {
	model = PopulateLensMetricDefaults(model)
	if model == nil {
		return model
	}
	if _, exists := model[attrVisible]; !exists {
		model[attrVisible] = true
	}
	if _, exists := model[attrAlignment]; !exists {
		model[attrAlignment] = attrAlignRight
	}
	return model
}
