/*
Connectors

OpenAPI schema for Connectors endpoints

API version: 0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package kibanaactions

import (
	"encoding/json"
)

// checks if the UpdateConnectorRequestServerlog type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateConnectorRequestServerlog{}

// UpdateConnectorRequestServerlog struct for UpdateConnectorRequestServerlog
type UpdateConnectorRequestServerlog struct {
	// The display name for the connector.
	Name string `json:"name"`
}

// NewUpdateConnectorRequestServerlog instantiates a new UpdateConnectorRequestServerlog object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateConnectorRequestServerlog(name string) *UpdateConnectorRequestServerlog {
	this := UpdateConnectorRequestServerlog{}
	this.Name = name
	return &this
}

// NewUpdateConnectorRequestServerlogWithDefaults instantiates a new UpdateConnectorRequestServerlog object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateConnectorRequestServerlogWithDefaults() *UpdateConnectorRequestServerlog {
	this := UpdateConnectorRequestServerlog{}
	return &this
}

// GetName returns the Name field value
func (o *UpdateConnectorRequestServerlog) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestServerlog) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UpdateConnectorRequestServerlog) SetName(v string) {
	o.Name = v
}

func (o UpdateConnectorRequestServerlog) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateConnectorRequestServerlog) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

type NullableUpdateConnectorRequestServerlog struct {
	value *UpdateConnectorRequestServerlog
	isSet bool
}

func (v NullableUpdateConnectorRequestServerlog) Get() *UpdateConnectorRequestServerlog {
	return v.value
}

func (v *NullableUpdateConnectorRequestServerlog) Set(val *UpdateConnectorRequestServerlog) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateConnectorRequestServerlog) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateConnectorRequestServerlog) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateConnectorRequestServerlog(val *UpdateConnectorRequestServerlog) *NullableUpdateConnectorRequestServerlog {
	return &NullableUpdateConnectorRequestServerlog{value: val, isSet: true}
}

func (v NullableUpdateConnectorRequestServerlog) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateConnectorRequestServerlog) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
