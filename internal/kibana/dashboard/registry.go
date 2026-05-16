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
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/esqlcontrol"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/image"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/markdown"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/optionslist"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/rangeslider"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloalerts"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloburnrate"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/sloerrorbudget"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/slooverview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticsmonitors"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/syntheticsstatsoverview"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/timeslider"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

// panelHandlers is populated as individual panel implementations land.
var panelHandlers = []iface.Handler{
	sloburnrate.Handler{},
	sloerrorbudget.Handler{},
	slooverview.Handler{},
	syntheticsmonitors.Handler{},
	syntheticsstatsoverview.Handler{},
	timeslider.Handler{},
	optionslist.Handler{},
	rangeslider.Handler{},
	esqlcontrol.Handler{},
	markdown.Handler{},
	image.Handler{},
	sloalerts.Handler{},
}

var panelTypeToHandler map[string]iface.Handler
var derivedPanelConfigNames []string

func init() {
	panelkit.GlobalPanelJSONDefaults = populatePanelConfigJSONDefaults
	panelTypeToHandler = make(map[string]iface.Handler, len(panelHandlers))
	derivedPanelConfigNames = append(derivedPanelConfigNames, "config_json")
	for _, h := range panelHandlers {
		pt := h.PanelType()
		if _, dup := panelTypeToHandler[pt]; dup {
			panic(fmt.Sprintf("dashboard: duplicate iface.Handler.PanelType registration: %q", pt))
		}
		panelTypeToHandler[pt] = h
		block := pt + "_config"
		panelkit.MustPanelConfigBlockTagged(block)
		derivedPanelConfigNames = append(derivedPanelConfigNames, block)
	}

	typedSiblings := append([]string{}, derivedPanelConfigNames...)
	typedSiblings = append(typedSiblings, []string{"vis_config", "lens_dashboard_app_config", "discover_session_config"}...)
	panelkit.SetTypedSiblingPanelConfigBlockNames(typedSiblings)
}

func LookupHandler(panelType string) iface.Handler { return panelTypeToHandler[panelType] }

func AllHandlers() []iface.Handler { return panelHandlers }

func ConfigNames() []string { return derivedPanelConfigNames }
