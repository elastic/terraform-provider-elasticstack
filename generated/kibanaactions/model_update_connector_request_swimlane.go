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

// checks if the UpdateConnectorRequestSwimlane type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateConnectorRequestSwimlane{}

// UpdateConnectorRequestSwimlane struct for UpdateConnectorRequestSwimlane
type UpdateConnectorRequestSwimlane struct {
	Config ConfigPropertiesSwimlane `json:"config"`
	// The display name for the connector.
	Name    string                    `json:"name"`
	Secrets SecretsPropertiesSwimlane `json:"secrets"`
}

// NewUpdateConnectorRequestSwimlane instantiates a new UpdateConnectorRequestSwimlane object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateConnectorRequestSwimlane(config ConfigPropertiesSwimlane, name string, secrets SecretsPropertiesSwimlane) *UpdateConnectorRequestSwimlane {
	this := UpdateConnectorRequestSwimlane{}
	this.Config = config
	this.Name = name
	this.Secrets = secrets
	return &this
}

// NewUpdateConnectorRequestSwimlaneWithDefaults instantiates a new UpdateConnectorRequestSwimlane object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateConnectorRequestSwimlaneWithDefaults() *UpdateConnectorRequestSwimlane {
	this := UpdateConnectorRequestSwimlane{}
	return &this
}

// GetConfig returns the Config field value
func (o *UpdateConnectorRequestSwimlane) GetConfig() ConfigPropertiesSwimlane {
	if o == nil {
		var ret ConfigPropertiesSwimlane
		return ret
	}

	return o.Config
}

// GetConfigOk returns a tuple with the Config field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestSwimlane) GetConfigOk() (*ConfigPropertiesSwimlane, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Config, true
}

// SetConfig sets field value
func (o *UpdateConnectorRequestSwimlane) SetConfig(v ConfigPropertiesSwimlane) {
	o.Config = v
}

// GetName returns the Name field value
func (o *UpdateConnectorRequestSwimlane) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestSwimlane) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UpdateConnectorRequestSwimlane) SetName(v string) {
	o.Name = v
}

// GetSecrets returns the Secrets field value
func (o *UpdateConnectorRequestSwimlane) GetSecrets() SecretsPropertiesSwimlane {
	if o == nil {
		var ret SecretsPropertiesSwimlane
		return ret
	}

	return o.Secrets
}

// GetSecretsOk returns a tuple with the Secrets field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestSwimlane) GetSecretsOk() (*SecretsPropertiesSwimlane, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Secrets, true
}

// SetSecrets sets field value
func (o *UpdateConnectorRequestSwimlane) SetSecrets(v SecretsPropertiesSwimlane) {
	o.Secrets = v
}

func (o UpdateConnectorRequestSwimlane) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateConnectorRequestSwimlane) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["config"] = o.Config
	toSerialize["name"] = o.Name
	toSerialize["secrets"] = o.Secrets
	return toSerialize, nil
}

type NullableUpdateConnectorRequestSwimlane struct {
	value *UpdateConnectorRequestSwimlane
	isSet bool
}

func (v NullableUpdateConnectorRequestSwimlane) Get() *UpdateConnectorRequestSwimlane {
	return v.value
}

func (v *NullableUpdateConnectorRequestSwimlane) Set(val *UpdateConnectorRequestSwimlane) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateConnectorRequestSwimlane) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateConnectorRequestSwimlane) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateConnectorRequestSwimlane(val *UpdateConnectorRequestSwimlane) *NullableUpdateConnectorRequestSwimlane {
	return &NullableUpdateConnectorRequestSwimlane{value: val, isSet: true}
}

func (v NullableUpdateConnectorRequestSwimlane) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateConnectorRequestSwimlane) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
