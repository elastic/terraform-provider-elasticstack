/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
)

// checks if the Objective type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Objective{}

// Objective Defines properties for the SLO objective
type Objective struct {
	// the target objective between 0 and 1 excluded
	Target float64 `json:"target"`
	// the target objective for each slice when using a timeslices budgeting method
	TimesliceTarget *float64 `json:"timesliceTarget,omitempty"`
	// the duration of each slice when using a timeslices budgeting method, as {duraton}{unit}
	TimesliceWindow *string `json:"timesliceWindow,omitempty"`
}

// NewObjective instantiates a new Objective object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewObjective(target float64) *Objective {
	this := Objective{}
	this.Target = target
	return &this
}

// NewObjectiveWithDefaults instantiates a new Objective object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewObjectiveWithDefaults() *Objective {
	this := Objective{}
	return &this
}

// GetTarget returns the Target field value
func (o *Objective) GetTarget() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Target
}

// GetTargetOk returns a tuple with the Target field value
// and a boolean to check if the value has been set.
func (o *Objective) GetTargetOk() (*float64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Target, true
}

// SetTarget sets field value
func (o *Objective) SetTarget(v float64) {
	o.Target = v
}

// GetTimesliceTarget returns the TimesliceTarget field value if set, zero value otherwise.
func (o *Objective) GetTimesliceTarget() float64 {
	if o == nil || IsNil(o.TimesliceTarget) {
		var ret float64
		return ret
	}
	return *o.TimesliceTarget
}

// GetTimesliceTargetOk returns a tuple with the TimesliceTarget field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Objective) GetTimesliceTargetOk() (*float64, bool) {
	if o == nil || IsNil(o.TimesliceTarget) {
		return nil, false
	}
	return o.TimesliceTarget, true
}

// HasTimesliceTarget returns a boolean if a field has been set.
func (o *Objective) HasTimesliceTarget() bool {
	if o != nil && !IsNil(o.TimesliceTarget) {
		return true
	}

	return false
}

// SetTimesliceTarget gets a reference to the given float64 and assigns it to the TimesliceTarget field.
func (o *Objective) SetTimesliceTarget(v float64) {
	o.TimesliceTarget = &v
}

// GetTimesliceWindow returns the TimesliceWindow field value if set, zero value otherwise.
func (o *Objective) GetTimesliceWindow() string {
	if o == nil || IsNil(o.TimesliceWindow) {
		var ret string
		return ret
	}
	return *o.TimesliceWindow
}

// GetTimesliceWindowOk returns a tuple with the TimesliceWindow field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Objective) GetTimesliceWindowOk() (*string, bool) {
	if o == nil || IsNil(o.TimesliceWindow) {
		return nil, false
	}
	return o.TimesliceWindow, true
}

// HasTimesliceWindow returns a boolean if a field has been set.
func (o *Objective) HasTimesliceWindow() bool {
	if o != nil && !IsNil(o.TimesliceWindow) {
		return true
	}

	return false
}

// SetTimesliceWindow gets a reference to the given string and assigns it to the TimesliceWindow field.
func (o *Objective) SetTimesliceWindow(v string) {
	o.TimesliceWindow = &v
}

func (o Objective) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Objective) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["target"] = o.Target
	if !IsNil(o.TimesliceTarget) {
		toSerialize["timesliceTarget"] = o.TimesliceTarget
	}
	if !IsNil(o.TimesliceWindow) {
		toSerialize["timesliceWindow"] = o.TimesliceWindow
	}
	return toSerialize, nil
}

type NullableObjective struct {
	value *Objective
	isSet bool
}

func (v NullableObjective) Get() *Objective {
	return v.value
}

func (v *NullableObjective) Set(val *Objective) {
	v.value = val
	v.isSet = true
}

func (v NullableObjective) IsSet() bool {
	return v.isSet
}

func (v *NullableObjective) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableObjective(val *Objective) *NullableObjective {
	return &NullableObjective{value: val, isSet: true}
}

func (v NullableObjective) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableObjective) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
