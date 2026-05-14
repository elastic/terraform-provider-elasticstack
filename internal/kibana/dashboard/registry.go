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
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloburnrate"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

// panelHandlers is populated as individual panel implementations land.
var panelHandlers = []iface.Handler{
	sloburnrate.Handler{},
}

var panelTypeToHandler map[string]iface.Handler
var derivedPanelConfigNames []string

func init() {
	sloburnrate.SetPopulatePanelConfigJSONDefaults(populatePanelConfigJSONDefaults)
	panelTypeToHandler = make(map[string]iface.Handler, len(panelHandlers))
	derivedPanelConfigNames = append(derivedPanelConfigNames, "config_json")
	for _, h := range panelHandlers {
		panelTypeToHandler[h.PanelType()] = h
		block := h.PanelType() + "_config"
		panelkit.MustPanelConfigBlockTagged(block)
		derivedPanelConfigNames = append(derivedPanelConfigNames, block)
	}
}

func LookupHandler(panelType string) iface.Handler { return panelTypeToHandler[panelType] }

func AllHandlers() []iface.Handler { return panelHandlers }

func ConfigNames() []string { return derivedPanelConfigNames }
