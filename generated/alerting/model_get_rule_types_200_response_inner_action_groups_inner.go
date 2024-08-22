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

// checks if the GetRuleTypes200ResponseInnerActionGroupsInner type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetRuleTypes200ResponseInnerActionGroupsInner{}

// GetRuleTypes200ResponseInnerActionGroupsInner struct for GetRuleTypes200ResponseInnerActionGroupsInner
type GetRuleTypes200ResponseInnerActionGroupsInner struct {
	Id   NullableString `json:"id,omitempty"`
	Name NullableString `json:"name,omitempty"`
}

// NewGetRuleTypes200ResponseInnerActionGroupsInner instantiates a new GetRuleTypes200ResponseInnerActionGroupsInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetRuleTypes200ResponseInnerActionGroupsInner() *GetRuleTypes200ResponseInnerActionGroupsInner {
	this := GetRuleTypes200ResponseInnerActionGroupsInner{}
	return &this
}

// NewGetRuleTypes200ResponseInnerActionGroupsInnerWithDefaults instantiates a new GetRuleTypes200ResponseInnerActionGroupsInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetRuleTypes200ResponseInnerActionGroupsInnerWithDefaults() *GetRuleTypes200ResponseInnerActionGroupsInner {
	this := GetRuleTypes200ResponseInnerActionGroupsInner{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) GetId() string {
	if o == nil || IsNil(o.Id.Get()) {
		var ret string
		return ret
	}
	return *o.Id.Get()
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Id.Get(), o.Id.IsSet()
}

// HasId returns a boolean if a field has been set.
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) HasId() bool {
	if o != nil && o.Id.IsSet() {
		return true
	}

	return false
}

// SetId gets a reference to the given NullableString and assigns it to the Id field.
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) SetId(v string) {
	o.Id.Set(&v)
}

// SetIdNil sets the value for Id to be an explicit nil
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) SetIdNil() {
	o.Id.Set(nil)
}

// UnsetId ensures that no value is present for Id, not even an explicit nil
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) UnsetId() {
	o.Id.Unset()
}

// GetName returns the Name field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) GetName() string {
	if o == nil || IsNil(o.Name.Get()) {
		var ret string
		return ret
	}
	return *o.Name.Get()
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Name.Get(), o.Name.IsSet()
}

// HasName returns a boolean if a field has been set.
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) HasName() bool {
	if o != nil && o.Name.IsSet() {
		return true
	}

	return false
}

// SetName gets a reference to the given NullableString and assigns it to the Name field.
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) SetName(v string) {
	o.Name.Set(&v)
}

// SetNameNil sets the value for Name to be an explicit nil
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) SetNameNil() {
	o.Name.Set(nil)
}

// UnsetName ensures that no value is present for Name, not even an explicit nil
func (o *GetRuleTypes200ResponseInnerActionGroupsInner) UnsetName() {
	o.Name.Unset()
}

func (o GetRuleTypes200ResponseInnerActionGroupsInner) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetRuleTypes200ResponseInnerActionGroupsInner) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if o.Id.IsSet() {
		toSerialize["id"] = o.Id.Get()
	}
	if o.Name.IsSet() {
		toSerialize["name"] = o.Name.Get()
	}
	return toSerialize, nil
}

type NullableGetRuleTypes200ResponseInnerActionGroupsInner struct {
	value *GetRuleTypes200ResponseInnerActionGroupsInner
	isSet bool
}

func (v NullableGetRuleTypes200ResponseInnerActionGroupsInner) Get() *GetRuleTypes200ResponseInnerActionGroupsInner {
	return v.value
}

func (v *NullableGetRuleTypes200ResponseInnerActionGroupsInner) Set(val *GetRuleTypes200ResponseInnerActionGroupsInner) {
	v.value = val
	v.isSet = true
}

func (v NullableGetRuleTypes200ResponseInnerActionGroupsInner) IsSet() bool {
	return v.isSet
}

func (v *NullableGetRuleTypes200ResponseInnerActionGroupsInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetRuleTypes200ResponseInnerActionGroupsInner(val *GetRuleTypes200ResponseInnerActionGroupsInner) *NullableGetRuleTypes200ResponseInnerActionGroupsInner {
	return &NullableGetRuleTypes200ResponseInnerActionGroupsInner{value: val, isSet: true}
}

func (v NullableGetRuleTypes200ResponseInnerActionGroupsInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetRuleTypes200ResponseInnerActionGroupsInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
