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

package models

import (
	"time"
)

type AlertingRule struct {
	RuleID     string
	SpaceID    string
	Name       string
	Consumer   string
	NotifyWhen *string
	Params     map[string]any
	RuleTypeID string
	Schedule   AlertingRuleSchedule
	Actions    []AlertingRuleAction
	Enabled    *bool
	Tags       []string
	Throttle   *string

	ScheduledTaskID *string
	ExecutionStatus AlertingRuleExecutionStatus
	AlertDelay      *float32
}

type AlertingRuleSchedule struct {
	Interval string
}

type AlertingRuleAction struct {
	Group        string
	ID           string
	Params       map[string]any
	Frequency    *ActionFrequency
	AlertsFilter *ActionAlertsFilter
}

type AlertingRuleExecutionStatus struct {
	LastExecutionDate *time.Time
	Status            *string
}

type ActionFrequency struct {
	Summary    bool
	NotifyWhen string
	Throttle   *string
}

type ActionAlertsFilter struct {
	Kql       *string
	Timeframe *AlertsFilterTimeframe
}

type AlertsFilterTimeframe struct {
	Days       []int32
	Timezone   string
	HoursStart string
	HoursEnd   string
}
