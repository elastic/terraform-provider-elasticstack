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

package panelkit

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/panel/iface"
)

// NoopHandlerBase is an embeddable zero-value struct that provides default no-op
// implementations of the four iface.Handler methods that are identical across all
// simple (non-Lens) panel types. Embed it in a panel handler struct and implement
// only the meaningful panel-specific methods.
type NoopHandlerBase struct{}

func (NoopHandlerBase) ClassifyJSON(_ map[string]any) bool                            { return false }
func (NoopHandlerBase) PopulateJSONDefaults(config map[string]any) map[string]any     { return config }
func (NoopHandlerBase) PinnedHandler() iface.PinnedHandler                            { return nil }
func (NoopHandlerBase) AlignStateFromPlan(_ context.Context, _, _ *models.PanelModel) {}
