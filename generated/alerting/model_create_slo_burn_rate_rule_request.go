/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.2
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
)

// checks if the CreateSloBurnRateRuleRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateSloBurnRateRuleRequest{}

// CreateSloBurnRateRuleRequest A rule that detects when the burn rate is above a defined threshold for two different lookback periods. The two periods are a long period and a short period that is 1/12th of the long period. For each lookback period, the burn rate is computed as the error rate divided by the error budget. When the burn rates for both periods surpass the threshold, an alert occurs.
type CreateSloBurnRateRuleRequest struct {
	Actions    []ActionsInner `json:"actions,omitempty"`
	AlertDelay *AlertDelay    `json:"alert_delay,omitempty"`
	// The name of the application or feature that owns the rule. For example: `alerts`, `apm`, `discover`, `infrastructure`, `logs`, `metrics`, `ml`, `monitoring`, `securitySolution`, `siem`, `stackAlerts`, or `uptime`.
	Consumer string `json:"consumer"`
	// Indicates whether you want to run the rule on an interval basis after it is created.
	Enabled *bool `json:"enabled,omitempty"`
	// The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.
	Name string `json:"name"`
	// Deprecated
	NotifyWhen *NotifyWhen               `json:"notify_when,omitempty"`
	Params     ParamsPropertySloBurnRate `json:"params"`
	// The ID of the rule type that you want to call when the rule is scheduled to run.
	RuleTypeId string   `json:"rule_type_id"`
	Schedule   Schedule `json:"schedule"`
	Tags       []string `json:"tags,omitempty"`
	// Deprecated in 8.13.0. Use the `throttle` property in the action `frequency` object instead. The throttle interval, which defines how often an alert generates repeated actions. NOTE: You cannot specify the throttle interval at both the rule and action level. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.
	// Deprecated
	Throttle NullableString `json:"throttle,omitempty"`
}

// NewCreateSloBurnRateRuleRequest instantiates a new CreateSloBurnRateRuleRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateSloBurnRateRuleRequest(consumer string, name string, params ParamsPropertySloBurnRate, ruleTypeId string, schedule Schedule) *CreateSloBurnRateRuleRequest {
	this := CreateSloBurnRateRuleRequest{}
	this.Consumer = consumer
	this.Name = name
	this.Params = params
	this.RuleTypeId = ruleTypeId
	this.Schedule = schedule
	return &this
}

// NewCreateSloBurnRateRuleRequestWithDefaults instantiates a new CreateSloBurnRateRuleRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateSloBurnRateRuleRequestWithDefaults() *CreateSloBurnRateRuleRequest {
	this := CreateSloBurnRateRuleRequest{}
	return &this
}

// GetActions returns the Actions field value if set, zero value otherwise.
func (o *CreateSloBurnRateRuleRequest) GetActions() []ActionsInner {
	if o == nil || IsNil(o.Actions) {
		var ret []ActionsInner
		return ret
	}
	return o.Actions
}

// GetActionsOk returns a tuple with the Actions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetActionsOk() ([]ActionsInner, bool) {
	if o == nil || IsNil(o.Actions) {
		return nil, false
	}
	return o.Actions, true
}

// HasActions returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasActions() bool {
	if o != nil && !IsNil(o.Actions) {
		return true
	}

	return false
}

// SetActions gets a reference to the given []ActionsInner and assigns it to the Actions field.
func (o *CreateSloBurnRateRuleRequest) SetActions(v []ActionsInner) {
	o.Actions = v
}

// GetAlertDelay returns the AlertDelay field value if set, zero value otherwise.
func (o *CreateSloBurnRateRuleRequest) GetAlertDelay() AlertDelay {
	if o == nil || IsNil(o.AlertDelay) {
		var ret AlertDelay
		return ret
	}
	return *o.AlertDelay
}

// GetAlertDelayOk returns a tuple with the AlertDelay field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetAlertDelayOk() (*AlertDelay, bool) {
	if o == nil || IsNil(o.AlertDelay) {
		return nil, false
	}
	return o.AlertDelay, true
}

// HasAlertDelay returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasAlertDelay() bool {
	if o != nil && !IsNil(o.AlertDelay) {
		return true
	}

	return false
}

// SetAlertDelay gets a reference to the given AlertDelay and assigns it to the AlertDelay field.
func (o *CreateSloBurnRateRuleRequest) SetAlertDelay(v AlertDelay) {
	o.AlertDelay = &v
}

// GetConsumer returns the Consumer field value
func (o *CreateSloBurnRateRuleRequest) GetConsumer() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Consumer
}

// GetConsumerOk returns a tuple with the Consumer field value
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetConsumerOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Consumer, true
}

// SetConsumer sets field value
func (o *CreateSloBurnRateRuleRequest) SetConsumer(v string) {
	o.Consumer = v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *CreateSloBurnRateRuleRequest) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *CreateSloBurnRateRuleRequest) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetName returns the Name field value
func (o *CreateSloBurnRateRuleRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateSloBurnRateRuleRequest) SetName(v string) {
	o.Name = v
}

// GetNotifyWhen returns the NotifyWhen field value if set, zero value otherwise.
// Deprecated
func (o *CreateSloBurnRateRuleRequest) GetNotifyWhen() NotifyWhen {
	if o == nil || IsNil(o.NotifyWhen) {
		var ret NotifyWhen
		return ret
	}
	return *o.NotifyWhen
}

// GetNotifyWhenOk returns a tuple with the NotifyWhen field value if set, nil otherwise
// and a boolean to check if the value has been set.
// Deprecated
func (o *CreateSloBurnRateRuleRequest) GetNotifyWhenOk() (*NotifyWhen, bool) {
	if o == nil || IsNil(o.NotifyWhen) {
		return nil, false
	}
	return o.NotifyWhen, true
}

// HasNotifyWhen returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasNotifyWhen() bool {
	if o != nil && !IsNil(o.NotifyWhen) {
		return true
	}

	return false
}

// SetNotifyWhen gets a reference to the given NotifyWhen and assigns it to the NotifyWhen field.
// Deprecated
func (o *CreateSloBurnRateRuleRequest) SetNotifyWhen(v NotifyWhen) {
	o.NotifyWhen = &v
}

// GetParams returns the Params field value
func (o *CreateSloBurnRateRuleRequest) GetParams() ParamsPropertySloBurnRate {
	if o == nil {
		var ret ParamsPropertySloBurnRate
		return ret
	}

	return o.Params
}

// GetParamsOk returns a tuple with the Params field value
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetParamsOk() (*ParamsPropertySloBurnRate, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Params, true
}

// SetParams sets field value
func (o *CreateSloBurnRateRuleRequest) SetParams(v ParamsPropertySloBurnRate) {
	o.Params = v
}

// GetRuleTypeId returns the RuleTypeId field value
func (o *CreateSloBurnRateRuleRequest) GetRuleTypeId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RuleTypeId
}

// GetRuleTypeIdOk returns a tuple with the RuleTypeId field value
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetRuleTypeIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RuleTypeId, true
}

// SetRuleTypeId sets field value
func (o *CreateSloBurnRateRuleRequest) SetRuleTypeId(v string) {
	o.RuleTypeId = v
}

// GetSchedule returns the Schedule field value
func (o *CreateSloBurnRateRuleRequest) GetSchedule() Schedule {
	if o == nil {
		var ret Schedule
		return ret
	}

	return o.Schedule
}

// GetScheduleOk returns a tuple with the Schedule field value
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetScheduleOk() (*Schedule, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Schedule, true
}

// SetSchedule sets field value
func (o *CreateSloBurnRateRuleRequest) SetSchedule(v Schedule) {
	o.Schedule = v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *CreateSloBurnRateRuleRequest) GetTags() []string {
	if o == nil || IsNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateSloBurnRateRuleRequest) GetTagsOk() ([]string, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *CreateSloBurnRateRuleRequest) SetTags(v []string) {
	o.Tags = v
}

// GetThrottle returns the Throttle field value if set, zero value otherwise (both if not set or set to explicit null).
// Deprecated
func (o *CreateSloBurnRateRuleRequest) GetThrottle() string {
	if o == nil || IsNil(o.Throttle.Get()) {
		var ret string
		return ret
	}
	return *o.Throttle.Get()
}

// GetThrottleOk returns a tuple with the Throttle field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
// Deprecated
func (o *CreateSloBurnRateRuleRequest) GetThrottleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Throttle.Get(), o.Throttle.IsSet()
}

// HasThrottle returns a boolean if a field has been set.
func (o *CreateSloBurnRateRuleRequest) HasThrottle() bool {
	if o != nil && o.Throttle.IsSet() {
		return true
	}

	return false
}

// SetThrottle gets a reference to the given NullableString and assigns it to the Throttle field.
// Deprecated
func (o *CreateSloBurnRateRuleRequest) SetThrottle(v string) {
	o.Throttle.Set(&v)
}

// SetThrottleNil sets the value for Throttle to be an explicit nil
func (o *CreateSloBurnRateRuleRequest) SetThrottleNil() {
	o.Throttle.Set(nil)
}

// UnsetThrottle ensures that no value is present for Throttle, not even an explicit nil
func (o *CreateSloBurnRateRuleRequest) UnsetThrottle() {
	o.Throttle.Unset()
}

func (o CreateSloBurnRateRuleRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateSloBurnRateRuleRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Actions) {
		toSerialize["actions"] = o.Actions
	}
	if !IsNil(o.AlertDelay) {
		toSerialize["alert_delay"] = o.AlertDelay
	}
	toSerialize["consumer"] = o.Consumer
	if !IsNil(o.Enabled) {
		toSerialize["enabled"] = o.Enabled
	}
	toSerialize["name"] = o.Name
	if !IsNil(o.NotifyWhen) {
		toSerialize["notify_when"] = o.NotifyWhen
	}
	toSerialize["params"] = o.Params
	toSerialize["rule_type_id"] = o.RuleTypeId
	toSerialize["schedule"] = o.Schedule
	if !IsNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}
	if o.Throttle.IsSet() {
		toSerialize["throttle"] = o.Throttle.Get()
	}
	return toSerialize, nil
}

type NullableCreateSloBurnRateRuleRequest struct {
	value *CreateSloBurnRateRuleRequest
	isSet bool
}

func (v NullableCreateSloBurnRateRuleRequest) Get() *CreateSloBurnRateRuleRequest {
	return v.value
}

func (v *NullableCreateSloBurnRateRuleRequest) Set(val *CreateSloBurnRateRuleRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateSloBurnRateRuleRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateSloBurnRateRuleRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateSloBurnRateRuleRequest(val *CreateSloBurnRateRuleRequest) *NullableCreateSloBurnRateRuleRequest {
	return &NullableCreateSloBurnRateRuleRequest{value: val, isSet: true}
}

func (v NullableCreateSloBurnRateRuleRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateSloBurnRateRuleRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
