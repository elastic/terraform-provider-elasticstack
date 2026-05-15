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

package image_test

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/image"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit/contracttest"
)

func TestContract_fileSrc(t *testing.T) {
	t.Parallel()
	contracttest.Run(t, image.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "image",
			"grid": {"x": 0, "y": 0, "w": 10, "h": 10},
			"id": "image-contract",
			"config": {
				"image_config": {
					"src": {"type": "file", "file_id": "file-contract-123"},
					"alt_text": "diagram"
				}
			}
		}`,
		OmitRequiredLeafPresence: true,
		OmitValidateRequiredZero: true,
		SkipFields: []string{
			"object_fit",
			"hide_title",
			"hide_border",
			"title",
			"description",
			"background_color",
			"src.file.file_id",
		},
	})
}

func TestContract_urlSrcWithDrilldowns(t *testing.T) {
	t.Parallel()
	contracttest.Run(t, image.Handler{}, contracttest.Config{
		FullAPIResponse: `{
			"type": "image",
			"grid": {"x": 1, "y": 1, "w": 12, "h": 8},
			"id": "image-drill-contract",
			"config": {
				"image_config": {
					"src": {"type": "url", "url": "https://example.com/a.png"}
				},
				"drilldowns": [{
					"type": "dashboard_drilldown",
					"trigger": "on_click_image",
					"dashboard_id": "dash-99",
					"label": "Open other"
				}]
			}
		}`,
		OmitRequiredLeafPresence: true,
		OmitValidateRequiredZero: true,
		SkipFields: []string{
			"object_fit",
			"hide_title",
			"hide_border",
			"src.url.url",
			"drilldowns",
			"config.drilldowns",
		},
	})
}
