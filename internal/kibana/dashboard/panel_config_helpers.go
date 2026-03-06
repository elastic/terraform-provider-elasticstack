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
	"encoding/json"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

func panelConfigMap(config json.RawMessage) (map[string]any, error) {
	var cfgMap map[string]any
	if len(config) == 0 || string(config) == "null" {
		return cfgMap, nil
	}
	err := json.Unmarshal(config, &cfgMap)
	return cfgMap, err
}

func panelConfigRawFromMap(configMap map[string]any) (json.RawMessage, error) {
	if configMap == nil {
		return json.RawMessage("null"), nil
	}
	b, err := json.Marshal(configMap)
	return json.RawMessage(b), err
}

func panelConfigRawFromLensAttributes(attrs kbapi.KbnDashboardPanelLensConfig0Attributes0) (json.RawMessage, error) {
	var configAttrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	if err := configAttrs.FromKbnDashboardPanelLensConfig0Attributes0(attrs); err != nil {
		return nil, err
	}

	config0 := kbapi.KbnDashboardPanelLensConfig0{
		Attributes: configAttrs,
	}

	var lensConfig kbapi.KbnDashboardPanelLens_Config
	if err := lensConfig.FromKbnDashboardPanelLensConfig0(config0); err != nil {
		return nil, err
	}

	b, err := lensConfig.MarshalJSON()
	return json.RawMessage(b), err
}
