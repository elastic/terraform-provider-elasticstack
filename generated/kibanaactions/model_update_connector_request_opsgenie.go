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

// checks if the UpdateConnectorRequestOpsgenie type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UpdateConnectorRequestOpsgenie{}

// UpdateConnectorRequestOpsgenie struct for UpdateConnectorRequestOpsgenie
type UpdateConnectorRequestOpsgenie struct {
	Config ConfigPropertiesOpsgenie `json:"config"`
	// The display name for the connector.
	Name    string                    `json:"name"`
	Secrets SecretsPropertiesOpsgenie `json:"secrets"`
}

// NewUpdateConnectorRequestOpsgenie instantiates a new UpdateConnectorRequestOpsgenie object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUpdateConnectorRequestOpsgenie(config ConfigPropertiesOpsgenie, name string, secrets SecretsPropertiesOpsgenie) *UpdateConnectorRequestOpsgenie {
	this := UpdateConnectorRequestOpsgenie{}
	this.Config = config
	this.Name = name
	this.Secrets = secrets
	return &this
}

// NewUpdateConnectorRequestOpsgenieWithDefaults instantiates a new UpdateConnectorRequestOpsgenie object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUpdateConnectorRequestOpsgenieWithDefaults() *UpdateConnectorRequestOpsgenie {
	this := UpdateConnectorRequestOpsgenie{}
	return &this
}

// GetConfig returns the Config field value
func (o *UpdateConnectorRequestOpsgenie) GetConfig() ConfigPropertiesOpsgenie {
	if o == nil {
		var ret ConfigPropertiesOpsgenie
		return ret
	}

	return o.Config
}

// GetConfigOk returns a tuple with the Config field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestOpsgenie) GetConfigOk() (*ConfigPropertiesOpsgenie, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Config, true
}

// SetConfig sets field value
func (o *UpdateConnectorRequestOpsgenie) SetConfig(v ConfigPropertiesOpsgenie) {
	o.Config = v
}

// GetName returns the Name field value
func (o *UpdateConnectorRequestOpsgenie) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestOpsgenie) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *UpdateConnectorRequestOpsgenie) SetName(v string) {
	o.Name = v
}

// GetSecrets returns the Secrets field value
func (o *UpdateConnectorRequestOpsgenie) GetSecrets() SecretsPropertiesOpsgenie {
	if o == nil {
		var ret SecretsPropertiesOpsgenie
		return ret
	}

	return o.Secrets
}

// GetSecretsOk returns a tuple with the Secrets field value
// and a boolean to check if the value has been set.
func (o *UpdateConnectorRequestOpsgenie) GetSecretsOk() (*SecretsPropertiesOpsgenie, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Secrets, true
}

// SetSecrets sets field value
func (o *UpdateConnectorRequestOpsgenie) SetSecrets(v SecretsPropertiesOpsgenie) {
	o.Secrets = v
}

func (o UpdateConnectorRequestOpsgenie) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UpdateConnectorRequestOpsgenie) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["config"] = o.Config
	toSerialize["name"] = o.Name
	toSerialize["secrets"] = o.Secrets
	return toSerialize, nil
}

type NullableUpdateConnectorRequestOpsgenie struct {
	value *UpdateConnectorRequestOpsgenie
	isSet bool
}

func (v NullableUpdateConnectorRequestOpsgenie) Get() *UpdateConnectorRequestOpsgenie {
	return v.value
}

func (v *NullableUpdateConnectorRequestOpsgenie) Set(val *UpdateConnectorRequestOpsgenie) {
	v.value = val
	v.isSet = true
}

func (v NullableUpdateConnectorRequestOpsgenie) IsSet() bool {
	return v.isSet
}

func (v *NullableUpdateConnectorRequestOpsgenie) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUpdateConnectorRequestOpsgenie(val *UpdateConnectorRequestOpsgenie) *NullableUpdateConnectorRequestOpsgenie {
	return &NullableUpdateConnectorRequestOpsgenie{value: val, isSet: true}
}

func (v NullableUpdateConnectorRequestOpsgenie) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUpdateConnectorRequestOpsgenie) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
