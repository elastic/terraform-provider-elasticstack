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

// checks if the SeverityMapping type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SeverityMapping{}

// SeverityMapping Mapping for the severity.
type SeverityMapping struct {
	// The type of field in Swimlane.
	FieldType string `json:"fieldType"`
	// The identifier for the field in Swimlane.
	Id string `json:"id"`
	// The key for the field in Swimlane.
	Key string `json:"key"`
	// The name of the field in Swimlane.
	Name string `json:"name"`
}

// NewSeverityMapping instantiates a new SeverityMapping object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSeverityMapping(fieldType string, id string, key string, name string) *SeverityMapping {
	this := SeverityMapping{}
	this.FieldType = fieldType
	this.Id = id
	this.Key = key
	this.Name = name
	return &this
}

// NewSeverityMappingWithDefaults instantiates a new SeverityMapping object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSeverityMappingWithDefaults() *SeverityMapping {
	this := SeverityMapping{}
	return &this
}

// GetFieldType returns the FieldType field value
func (o *SeverityMapping) GetFieldType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.FieldType
}

// GetFieldTypeOk returns a tuple with the FieldType field value
// and a boolean to check if the value has been set.
func (o *SeverityMapping) GetFieldTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.FieldType, true
}

// SetFieldType sets field value
func (o *SeverityMapping) SetFieldType(v string) {
	o.FieldType = v
}

// GetId returns the Id field value
func (o *SeverityMapping) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *SeverityMapping) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *SeverityMapping) SetId(v string) {
	o.Id = v
}

// GetKey returns the Key field value
func (o *SeverityMapping) GetKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Key
}

// GetKeyOk returns a tuple with the Key field value
// and a boolean to check if the value has been set.
func (o *SeverityMapping) GetKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Key, true
}

// SetKey sets field value
func (o *SeverityMapping) SetKey(v string) {
	o.Key = v
}

// GetName returns the Name field value
func (o *SeverityMapping) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *SeverityMapping) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *SeverityMapping) SetName(v string) {
	o.Name = v
}

func (o SeverityMapping) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SeverityMapping) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["fieldType"] = o.FieldType
	toSerialize["id"] = o.Id
	toSerialize["key"] = o.Key
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

type NullableSeverityMapping struct {
	value *SeverityMapping
	isSet bool
}

func (v NullableSeverityMapping) Get() *SeverityMapping {
	return v.value
}

func (v *NullableSeverityMapping) Set(val *SeverityMapping) {
	v.value = val
	v.isSet = true
}

func (v NullableSeverityMapping) IsSet() bool {
	return v.isSet
}

func (v *NullableSeverityMapping) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSeverityMapping(val *SeverityMapping) *NullableSeverityMapping {
	return &NullableSeverityMapping{value: val, isSet: true}
}

func (v NullableSeverityMapping) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSeverityMapping) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}