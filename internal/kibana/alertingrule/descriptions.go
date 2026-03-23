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

package alertingrule

import _ "embed"

//go:embed descriptions/notify_when.md
var notifyWhenDescription string

//go:embed descriptions/rule_type_id.md
var ruleTypeIDDescription string

//go:embed descriptions/throttle_rule.md
var throttleRuleDescription string

//go:embed descriptions/actions_group.md
var actionsGroupDescription string

//go:embed descriptions/actions_frequency.md
var actionsFrequencyDescription string

//go:embed descriptions/actions_frequency_notify_when.md
var actionsFrequencyNotifyWhenDescription string

//go:embed descriptions/actions_frequency_throttle.md
var actionsFrequencyThrottleDescription string

//go:embed descriptions/alerts_filter.md
var alertsFilterDescription string

//go:embed descriptions/timeframe_days.md
var timeframeDaysDescription string
