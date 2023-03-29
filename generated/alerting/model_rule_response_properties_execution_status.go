/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
	"time"
)

// checks if the RuleResponsePropertiesExecutionStatus type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RuleResponsePropertiesExecutionStatus{}

// RuleResponsePropertiesExecutionStatus struct for RuleResponsePropertiesExecutionStatus
type RuleResponsePropertiesExecutionStatus struct {
	LastDuration      *int32     `json:"last_duration,omitempty"`
	LastExecutionDate *time.Time `json:"last_execution_date,omitempty"`
	Status            *string    `json:"status,omitempty"`
}

// NewRuleResponsePropertiesExecutionStatus instantiates a new RuleResponsePropertiesExecutionStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRuleResponsePropertiesExecutionStatus() *RuleResponsePropertiesExecutionStatus {
	this := RuleResponsePropertiesExecutionStatus{}
	return &this
}

// NewRuleResponsePropertiesExecutionStatusWithDefaults instantiates a new RuleResponsePropertiesExecutionStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRuleResponsePropertiesExecutionStatusWithDefaults() *RuleResponsePropertiesExecutionStatus {
	this := RuleResponsePropertiesExecutionStatus{}
	return &this
}

// GetLastDuration returns the LastDuration field value if set, zero value otherwise.
func (o *RuleResponsePropertiesExecutionStatus) GetLastDuration() int32 {
	if o == nil || IsNil(o.LastDuration) {
		var ret int32
		return ret
	}
	return *o.LastDuration
}

// GetLastDurationOk returns a tuple with the LastDuration field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RuleResponsePropertiesExecutionStatus) GetLastDurationOk() (*int32, bool) {
	if o == nil || IsNil(o.LastDuration) {
		return nil, false
	}
	return o.LastDuration, true
}

// HasLastDuration returns a boolean if a field has been set.
func (o *RuleResponsePropertiesExecutionStatus) HasLastDuration() bool {
	if o != nil && !IsNil(o.LastDuration) {
		return true
	}

	return false
}

// SetLastDuration gets a reference to the given int32 and assigns it to the LastDuration field.
func (o *RuleResponsePropertiesExecutionStatus) SetLastDuration(v int32) {
	o.LastDuration = &v
}

// GetLastExecutionDate returns the LastExecutionDate field value if set, zero value otherwise.
func (o *RuleResponsePropertiesExecutionStatus) GetLastExecutionDate() time.Time {
	if o == nil || IsNil(o.LastExecutionDate) {
		var ret time.Time
		return ret
	}
	return *o.LastExecutionDate
}

// GetLastExecutionDateOk returns a tuple with the LastExecutionDate field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RuleResponsePropertiesExecutionStatus) GetLastExecutionDateOk() (*time.Time, bool) {
	if o == nil || IsNil(o.LastExecutionDate) {
		return nil, false
	}
	return o.LastExecutionDate, true
}

// HasLastExecutionDate returns a boolean if a field has been set.
func (o *RuleResponsePropertiesExecutionStatus) HasLastExecutionDate() bool {
	if o != nil && !IsNil(o.LastExecutionDate) {
		return true
	}

	return false
}

// SetLastExecutionDate gets a reference to the given time.Time and assigns it to the LastExecutionDate field.
func (o *RuleResponsePropertiesExecutionStatus) SetLastExecutionDate(v time.Time) {
	o.LastExecutionDate = &v
}

// GetStatus returns the Status field value if set, zero value otherwise.
func (o *RuleResponsePropertiesExecutionStatus) GetStatus() string {
	if o == nil || IsNil(o.Status) {
		var ret string
		return ret
	}
	return *o.Status
}

// GetStatusOk returns a tuple with the Status field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *RuleResponsePropertiesExecutionStatus) GetStatusOk() (*string, bool) {
	if o == nil || IsNil(o.Status) {
		return nil, false
	}
	return o.Status, true
}

// HasStatus returns a boolean if a field has been set.
func (o *RuleResponsePropertiesExecutionStatus) HasStatus() bool {
	if o != nil && !IsNil(o.Status) {
		return true
	}

	return false
}

// SetStatus gets a reference to the given string and assigns it to the Status field.
func (o *RuleResponsePropertiesExecutionStatus) SetStatus(v string) {
	o.Status = &v
}

func (o RuleResponsePropertiesExecutionStatus) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RuleResponsePropertiesExecutionStatus) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.LastDuration) {
		toSerialize["last_duration"] = o.LastDuration
	}
	if !IsNil(o.LastExecutionDate) {
		toSerialize["last_execution_date"] = o.LastExecutionDate
	}
	if !IsNil(o.Status) {
		toSerialize["status"] = o.Status
	}
	return toSerialize, nil
}

type NullableRuleResponsePropertiesExecutionStatus struct {
	value *RuleResponsePropertiesExecutionStatus
	isSet bool
}

func (v NullableRuleResponsePropertiesExecutionStatus) Get() *RuleResponsePropertiesExecutionStatus {
	return v.value
}

func (v *NullableRuleResponsePropertiesExecutionStatus) Set(val *RuleResponsePropertiesExecutionStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableRuleResponsePropertiesExecutionStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableRuleResponsePropertiesExecutionStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRuleResponsePropertiesExecutionStatus(val *RuleResponsePropertiesExecutionStatus) *NullableRuleResponsePropertiesExecutionStatus {
	return &NullableRuleResponsePropertiesExecutionStatus{value: val, isSet: true}
}

func (v NullableRuleResponsePropertiesExecutionStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRuleResponsePropertiesExecutionStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
