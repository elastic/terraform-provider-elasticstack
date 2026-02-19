package alerting_rule

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
