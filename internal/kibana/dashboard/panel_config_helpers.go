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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
)

type apiPanelConfig struct {
	Lens     *kbapi.KbnDashboardPanelLens_Config
	Markdown *kbapi.KbnDashboardPanelDASHBOARDMARKDOWN_Config
}

func (c apiPanelConfig) jsonString() (string, error) {
	switch {
	case c.Lens != nil:
		b, err := c.Lens.MarshalJSON()
		return string(b), err
	case c.Markdown != nil:
		b, err := c.Markdown.MarshalJSON()
		return string(b), err
	default:
		return "null", nil
	}
}

func panelConfigMap(config apiPanelConfig) (map[string]any, error) {
	var cfgMap map[string]any
	configJSON, err := config.jsonString()
	if err != nil {
		return nil, err
	}
	if configJSON == "" || configJSON == "null" {
		return cfgMap, nil
	}
	err = json.Unmarshal([]byte(configJSON), &cfgMap)
	return cfgMap, err
}

func panelConfigFromMap(panelType string, configMap map[string]any) (apiPanelConfig, error) {
	if configMap == nil {
		return apiPanelConfig{}, nil
	}
	b, err := json.Marshal(configMap)
	if err != nil {
		return apiPanelConfig{}, err
	}

	switch panelType {
	case "lens":
		var cfg kbapi.KbnDashboardPanelLens_Config
		if err := cfg.UnmarshalJSON(b); err != nil {
			return apiPanelConfig{}, err
		}
		return apiPanelConfig{Lens: &cfg}, nil
	case "DASHBOARD_MARKDOWN":
		var cfg kbapi.KbnDashboardPanelDASHBOARDMARKDOWN_Config
		if err := cfg.UnmarshalJSON(b); err != nil {
			return apiPanelConfig{}, err
		}
		return apiPanelConfig{Markdown: &cfg}, nil
	default:
		return apiPanelConfig{}, fmt.Errorf("unsupported panel type %q", panelType)
	}
}

func panelConfigFromLensAttributes(attrs kbapi.KbnDashboardPanelLensConfig0Attributes0) (apiPanelConfig, error) {
	var configAttrs kbapi.KbnDashboardPanelLens_Config_0_Attributes
	if err := configAttrs.FromKbnDashboardPanelLensConfig0Attributes0(attrs); err != nil {
		return apiPanelConfig{}, err
	}

	config0 := kbapi.KbnDashboardPanelLensConfig0{
		Attributes: configAttrs,
	}

	var lensConfig kbapi.KbnDashboardPanelLens_Config
	if err := lensConfig.FromKbnDashboardPanelLensConfig0(config0); err != nil {
		return apiPanelConfig{}, err
	}

	return apiPanelConfig{Lens: &lensConfig}, nil
}
