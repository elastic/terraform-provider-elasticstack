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

package dashboard

import (
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/lenscommon"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
)

// populatePanelConfigJSONDefaults normalizes panel config_json for semantic equality.
//
// Typed panel handlers may claim opaque JSON blobs via ClassifyJSON;
// PopulateJSONDefaults applies normalization for plan-time semantic equality (#1789).
//
// Lens-by-value panels still serialize as opaque "attributes"; that tree is unchanged
// here until dashboard-lens-contract introduces a Lens handler claiming that shape.
func populatePanelConfigJSONDefaults(config map[string]any) map[string]any {
	return populatePanelConfigJSONDefaultsWithHandlers(config, AllHandlers())
}

func populatePanelConfigJSONDefaultsWithHandlers(config map[string]any, handlers []iface.Handler) map[string]any {
	if config == nil {
		return config
	}

	for _, h := range handlers {
		if h.ClassifyJSON(config) {
			config = h.PopulateJSONDefaults(config)
			break
		}
	}

	// Lens visualization config_json normalization (opaque attributes.* structure).
	if attrs, ok := config["attributes"].(map[string]any); ok {
		config["attributes"] = populateLensAttributesDefaults(attrs)
	}

	return config
}

// populateLensAttributesDefaults applies type-specific defaults to lens attributes via registered VizConverters.
func populateLensAttributesDefaults(attrs map[string]any) map[string]any {
	if attrs == nil {
		return attrs
	}
	visType, _ := attrs["type"].(string)
	if c := lenscommon.ForType(visType); c != nil {
		return c.PopulateJSONDefaults(attrs)
	}
	return attrs
}
