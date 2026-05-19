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

package iface

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

type enclosingDashboardCtxKey struct{}

// WithEnclosingDashboard attaches the dashboard model under construction so handlers can read
// dashboard-level defaults during FromAPI (for example vis Lens presentation inheritance).
func WithEnclosingDashboard(ctx context.Context, dm *models.DashboardModel) context.Context {
	if dm == nil {
		return ctx
	}
	return context.WithValue(ctx, enclosingDashboardCtxKey{}, dm)
}

// EnclosingDashboard returns the dashboard passed via WithEnclosingDashboard, or nil.
func EnclosingDashboard(ctx context.Context) *models.DashboardModel {
	v, _ := ctx.Value(enclosingDashboardCtxKey{}).(*models.DashboardModel)
	return v
}
