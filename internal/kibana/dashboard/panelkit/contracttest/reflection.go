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

package contracttest

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panelkit"
)

func appendReflectIssues(ctx context.Context, handler iface.Handler, fixture string, issues *[]string) {
	block := handler.PanelType() + "_config"
	if !panelkit.HasPanelConfigBlock(block) {
		return
	}

	item0, err := ParseDashboardPanel(fixture)
	if err != nil {
		*issues = append(*issues, fmt.Sprintf("[Reflect] parse: %v", err))
		return
	}

	var pm models.PanelModel
	if diags := handler.FromAPI(ctx, &pm, nil, item0); diags.HasError() {
		*issues = append(*issues, fmt.Sprintf("[Reflect] FromAPI: %s", summarizeDiags(diags)))
		return
	}

	if !panelkit.HasConfig(&pm, block) {
		*issues = append(*issues, fmt.Sprintf("[Reflect] expected HasConfig(%s) after FromAPI", block))
		return
	}
	panelkit.ClearConfig(&pm, block)
	if panelkit.HasConfig(&pm, block) {
		*issues = append(*issues, fmt.Sprintf("[Reflect] expected HasConfig(%s) false after ClearConfig", block))
		return
	}
}
