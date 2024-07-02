/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
	"fmt"
)

// NotifyWhenAction Indicates how often alerts generate actions. Valid values include: `onActionGroupChange`: Actions run when the alert status changes; `onActiveAlert`: Actions run when the alert becomes active and at each check interval while the rule conditions are met; `onThrottleInterval`: Actions run when the alert becomes active and at the interval specified in the throttle property while the rule conditions are met. NOTE: You cannot specify `notify_when` at both the rule and action level. The recommended method is to set it for each action. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.
type NotifyWhenAction string

// List of notify_when_action
const (
	NOTIFY_WHEN_ACTION_ON_ACTION_GROUP_CHANGE NotifyWhenAction = "onActionGroupChange"
	NOTIFY_WHEN_ACTION_ON_ACTIVE_ALERT        NotifyWhenAction = "onActiveAlert"
	NOTIFY_WHEN_ACTION_ON_THROTTLE_INTERVAL   NotifyWhenAction = "onThrottleInterval"
)

// All allowed values of NotifyWhenAction enum
var AllowedNotifyWhenActionEnumValues = []NotifyWhenAction{
	"onActionGroupChange",
	"onActiveAlert",
	"onThrottleInterval",
}

func (v *NotifyWhenAction) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := NotifyWhenAction(value)
	for _, existing := range AllowedNotifyWhenActionEnumValues {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid NotifyWhenAction", value)
}

// NewNotifyWhenActionFromValue returns a pointer to a valid NotifyWhenAction
// for the value passed as argument, or an error if the value passed is not allowed by the enum
func NewNotifyWhenActionFromValue(v string) (*NotifyWhenAction, error) {
	ev := NotifyWhenAction(v)
	if ev.IsValid() {
		return &ev, nil
	} else {
		return nil, fmt.Errorf("invalid value '%v' for NotifyWhenAction: valid values are %v", v, AllowedNotifyWhenActionEnumValues)
	}
}

// IsValid return true if the value is valid for the enum, false otherwise
func (v NotifyWhenAction) IsValid() bool {
	for _, existing := range AllowedNotifyWhenActionEnumValues {
		if existing == v {
			return true
		}
	}
	return false
}

// Ptr returns reference to notify_when_action value
func (v NotifyWhenAction) Ptr() *NotifyWhenAction {
	return &v
}

type NullableNotifyWhenAction struct {
	value *NotifyWhenAction
	isSet bool
}

func (v NullableNotifyWhenAction) Get() *NotifyWhenAction {
	return v.value
}

func (v *NullableNotifyWhenAction) Set(val *NotifyWhenAction) {
	v.value = val
	v.isSet = true
}

func (v NullableNotifyWhenAction) IsSet() bool {
	return v.isSet
}

func (v *NullableNotifyWhenAction) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNotifyWhenAction(val *NotifyWhenAction) *NullableNotifyWhenAction {
	return &NullableNotifyWhenAction{value: val, isSet: true}
}

func (v NullableNotifyWhenAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNotifyWhenAction) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
