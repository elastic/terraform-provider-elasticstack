/*
Alerting

OpenAPI schema for alerting endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alerting

import (
	"encoding/json"
)

// checks if the GetRuleTypes200ResponseInnerActionVariablesParamsInner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetRuleTypes200ResponseInnerActionVariablesParamsInner{}

// GetRuleTypes200ResponseInnerActionVariablesParamsInner struct for GetRuleTypes200ResponseInnerActionVariablesParamsInner
type GetRuleTypes200ResponseInnerActionVariablesParamsInner struct {
	Description *string `json:"description,omitempty"`
	Name        *string `json:"name,omitempty"`
}

// NewGetRuleTypes200ResponseInnerActionVariablesParamsInner instantiates a new GetRuleTypes200ResponseInnerActionVariablesParamsInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetRuleTypes200ResponseInnerActionVariablesParamsInner() *GetRuleTypes200ResponseInnerActionVariablesParamsInner {
	this := GetRuleTypes200ResponseInnerActionVariablesParamsInner{}
	return &this
}

// NewGetRuleTypes200ResponseInnerActionVariablesParamsInnerWithDefaults instantiates a new GetRuleTypes200ResponseInnerActionVariablesParamsInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetRuleTypes200ResponseInnerActionVariablesParamsInnerWithDefaults() *GetRuleTypes200ResponseInnerActionVariablesParamsInner {
	this := GetRuleTypes200ResponseInnerActionVariablesParamsInner{}
	return &this
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) SetDescription(v string) {
	o.Description = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) GetName() string {
	if o == nil || IsNil(o.Name) {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) GetNameOk() (*string, bool) {
	if o == nil || IsNil(o.Name) {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) HasName() bool {
	if o != nil && !IsNil(o.Name) {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *GetRuleTypes200ResponseInnerActionVariablesParamsInner) SetName(v string) {
	o.Name = &v
}

func (o GetRuleTypes200ResponseInnerActionVariablesParamsInner) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetRuleTypes200ResponseInnerActionVariablesParamsInner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.Name) {
		toSerialize["name"] = o.Name
	}
	return toSerialize, nil
}

type NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner struct {
	value *GetRuleTypes200ResponseInnerActionVariablesParamsInner
	isSet bool
}

func (v NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) Get() *GetRuleTypes200ResponseInnerActionVariablesParamsInner {
	return v.value
}

func (v *NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) Set(val *GetRuleTypes200ResponseInnerActionVariablesParamsInner) {
	v.value = val
	v.isSet = true
}

func (v NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) IsSet() bool {
	return v.isSet
}

func (v *NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetRuleTypes200ResponseInnerActionVariablesParamsInner(val *GetRuleTypes200ResponseInnerActionVariablesParamsInner) *NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner {
	return &NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner{value: val, isSet: true}
}

func (v NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetRuleTypes200ResponseInnerActionVariablesParamsInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
