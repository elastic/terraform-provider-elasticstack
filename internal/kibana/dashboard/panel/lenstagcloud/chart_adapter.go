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

package lenstagcloud

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

func tagcloudChartFromLensAPI(config kbapi.KibanaHTTPAPIsLensApiConfig, destination any) error {
	raw, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal Lens chart: %w", err)
	}
	if err := json.Unmarshal(raw, destination); err != nil {
		return fmt.Errorf("decode tagcloud chart: %w", err)
	}
	return nil
}

func tagcloudChartToLensAPI(chart any) (kbapi.KibanaHTTPAPIsLensApiConfig, error) {
	raw, err := json.Marshal(chart)
	if err != nil {
		return kbapi.KibanaHTTPAPIsLensApiConfig{}, fmt.Errorf("marshal tagcloud chart: %w", err)
	}
	var result kbapi.KibanaHTTPAPIsLensApiConfig
	if err := json.Unmarshal(raw, &result); err != nil {
		return kbapi.KibanaHTTPAPIsLensApiConfig{}, fmt.Errorf("decode Lens chart: %w", err)
	}
	return result, nil
}
