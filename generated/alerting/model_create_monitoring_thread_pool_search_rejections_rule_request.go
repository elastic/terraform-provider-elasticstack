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

// checks if the CreateMonitoringThreadPoolSearchRejectionsRuleRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateMonitoringThreadPoolSearchRejectionsRuleRequest{}

// CreateMonitoringThreadPoolSearchRejectionsRuleRequest A rule that detects when the number of rejections in the thread pool exceeds a threshold.
type CreateMonitoringThreadPoolSearchRejectionsRuleRequest struct {
	Actions []ActionsInner `json:"actions,omitempty"`
	// The name of the application or feature that owns the rule. For example: `alerts`, `apm`, `discover`, `infrastructure`, `logs`, `metrics`, `ml`, `monitoring`, `securitySolution`, `siem`, `stackAlerts`, or `uptime`.
	Consumer string `json:"consumer"`
	// Indicates whether you want to run the rule on an interval basis after it is created.
	Enabled *bool `json:"enabled,omitempty"`
	// The name of the rule. While this name does not have to be unique, a distinctive name can help you identify a rule.
	Name string `json:"name"`
	// Deprecated
	NotifyWhen *NotifyWhen `json:"notify_when,omitempty"`
	// The parameters for a thread pool search rejections rule.
	Params map[string]interface{} `json:"params"`
	// The ID of the rule type that you want to call when the rule is scheduled to run.
	RuleTypeId string   `json:"rule_type_id"`
	Schedule   Schedule `json:"schedule"`
	// The tags for the rule.
	Tags []string `json:"tags,omitempty"`
	// Deprecated in 8.13.0. Use the `throttle` property in the action `frequency` object instead. The throttle interval, which defines how often an alert generates repeated actions. NOTE: You cannot specify the throttle interval at both the rule and action level. If you set it at the rule level then update the rule in Kibana, it is automatically changed to use action-specific values.
	// Deprecated
	Throttle interface{} `json:"throttle,omitempty"`
}

// NewCreateMonitoringThreadPoolSearchRejectionsRuleRequest instantiates a new CreateMonitoringThreadPoolSearchRejectionsRuleRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateMonitoringThreadPoolSearchRejectionsRuleRequest(consumer string, name string, params map[string]interface{}, ruleTypeId string, schedule Schedule) *CreateMonitoringThreadPoolSearchRejectionsRuleRequest {
	this := CreateMonitoringThreadPoolSearchRejectionsRuleRequest{}
	this.Consumer = consumer
	this.Name = name
	this.Params = params
	this.RuleTypeId = ruleTypeId
	this.Schedule = schedule
	return &this
}

// NewCreateMonitoringThreadPoolSearchRejectionsRuleRequestWithDefaults instantiates a new CreateMonitoringThreadPoolSearchRejectionsRuleRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateMonitoringThreadPoolSearchRejectionsRuleRequestWithDefaults() *CreateMonitoringThreadPoolSearchRejectionsRuleRequest {
	this := CreateMonitoringThreadPoolSearchRejectionsRuleRequest{}
	return &this
}

// GetActions returns the Actions field value if set, zero value otherwise.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetActions() []ActionsInner {
	if o == nil || IsNil(o.Actions) {
		var ret []ActionsInner
		return ret
	}
	return o.Actions
}

// GetActionsOk returns a tuple with the Actions field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetActionsOk() ([]ActionsInner, bool) {
	if o == nil || IsNil(o.Actions) {
		return nil, false
	}
	return o.Actions, true
}

// HasActions returns a boolean if a field has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) HasActions() bool {
	if o != nil && !IsNil(o.Actions) {
		return true
	}

	return false
}

// SetActions gets a reference to the given []ActionsInner and assigns it to the Actions field.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetActions(v []ActionsInner) {
	o.Actions = v
}

// GetConsumer returns the Consumer field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetConsumer() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Consumer
}

// GetConsumerOk returns a tuple with the Consumer field value
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetConsumerOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Consumer, true
}

// SetConsumer sets field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetConsumer(v string) {
	o.Consumer = v
}

// GetEnabled returns the Enabled field value if set, zero value otherwise.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetEnabled() bool {
	if o == nil || IsNil(o.Enabled) {
		var ret bool
		return ret
	}
	return *o.Enabled
}

// GetEnabledOk returns a tuple with the Enabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetEnabledOk() (*bool, bool) {
	if o == nil || IsNil(o.Enabled) {
		return nil, false
	}
	return o.Enabled, true
}

// HasEnabled returns a boolean if a field has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) HasEnabled() bool {
	if o != nil && !IsNil(o.Enabled) {
		return true
	}

	return false
}

// SetEnabled gets a reference to the given bool and assigns it to the Enabled field.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetEnabled(v bool) {
	o.Enabled = &v
}

// GetName returns the Name field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetName(v string) {
	o.Name = v
}

// GetNotifyWhen returns the NotifyWhen field value if set, zero value otherwise.
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetNotifyWhen() NotifyWhen {
	if o == nil || IsNil(o.NotifyWhen) {
		var ret NotifyWhen
		return ret
	}
	return *o.NotifyWhen
}

// GetNotifyWhenOk returns a tuple with the NotifyWhen field value if set, nil otherwise
// and a boolean to check if the value has been set.
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetNotifyWhenOk() (*NotifyWhen, bool) {
	if o == nil || IsNil(o.NotifyWhen) {
		return nil, false
	}
	return o.NotifyWhen, true
}

// HasNotifyWhen returns a boolean if a field has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) HasNotifyWhen() bool {
	if o != nil && !IsNil(o.NotifyWhen) {
		return true
	}

	return false
}

// SetNotifyWhen gets a reference to the given NotifyWhen and assigns it to the NotifyWhen field.
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetNotifyWhen(v NotifyWhen) {
	o.NotifyWhen = &v
}

// GetParams returns the Params field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetParams() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Params
}

// GetParamsOk returns a tuple with the Params field value
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetParamsOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Params, true
}

// SetParams sets field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetParams(v map[string]interface{}) {
	o.Params = v
}

// GetRuleTypeId returns the RuleTypeId field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetRuleTypeId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.RuleTypeId
}

// GetRuleTypeIdOk returns a tuple with the RuleTypeId field value
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetRuleTypeIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.RuleTypeId, true
}

// SetRuleTypeId sets field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetRuleTypeId(v string) {
	o.RuleTypeId = v
}

// GetSchedule returns the Schedule field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetSchedule() Schedule {
	if o == nil {
		var ret Schedule
		return ret
	}

	return o.Schedule
}

// GetScheduleOk returns a tuple with the Schedule field value
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetScheduleOk() (*Schedule, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Schedule, true
}

// SetSchedule sets field value
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetSchedule(v Schedule) {
	o.Schedule = v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetTags() []string {
	if o == nil || IsNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetTagsOk() ([]string, bool) {
	if o == nil || IsNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) HasTags() bool {
	if o != nil && !IsNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetTags(v []string) {
	o.Tags = v
}

// GetThrottle returns the Throttle field value if set, zero value otherwise (both if not set or set to explicit null).
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetThrottle() interface{} {
	if o == nil {
		var ret interface{}
		return ret
	}
	return o.Throttle
}

// GetThrottleOk returns a tuple with the Throttle field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) GetThrottleOk() (*interface{}, bool) {
	if o == nil || IsNil(o.Throttle) {
		return nil, false
	}
	return &o.Throttle, true
}

// HasThrottle returns a boolean if a field has been set.
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) HasThrottle() bool {
	if o != nil && IsNil(o.Throttle) {
		return true
	}

	return false
}

// SetThrottle gets a reference to the given interface{} and assigns it to the Throttle field.
// Deprecated
func (o *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) SetThrottle(v interface{}) {
	o.Throttle = v
}

func (o CreateMonitoringThreadPoolSearchRejectionsRuleRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateMonitoringThreadPoolSearchRejectionsRuleRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Actions) {
		toSerialize["actions"] = o.Actions
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
	if o.Throttle != nil {
		toSerialize["throttle"] = o.Throttle
	}
	return toSerialize, nil
}

type NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest struct {
	value *CreateMonitoringThreadPoolSearchRejectionsRuleRequest
	isSet bool
}

func (v NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) Get() *CreateMonitoringThreadPoolSearchRejectionsRuleRequest {
	return v.value
}

func (v *NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) Set(val *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest(val *CreateMonitoringThreadPoolSearchRejectionsRuleRequest) *NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest {
	return &NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest{value: val, isSet: true}
}

func (v NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateMonitoringThreadPoolSearchRejectionsRuleRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
