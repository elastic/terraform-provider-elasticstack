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

// checks if the SecretsPropertiesResilient type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SecretsPropertiesResilient{}

// SecretsPropertiesResilient Defines secrets for connectors when type is `.resilient`.
type SecretsPropertiesResilient struct {
	// The authentication key ID for HTTP Basic authentication.
	ApiKeyId string `json:"apiKeyId"`
	// The authentication key secret for HTTP Basic authentication.
	ApiKeySecret string `json:"apiKeySecret"`
}

// NewSecretsPropertiesResilient instantiates a new SecretsPropertiesResilient object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSecretsPropertiesResilient(apiKeyId string, apiKeySecret string) *SecretsPropertiesResilient {
	this := SecretsPropertiesResilient{}
	this.ApiKeyId = apiKeyId
	this.ApiKeySecret = apiKeySecret
	return &this
}

// NewSecretsPropertiesResilientWithDefaults instantiates a new SecretsPropertiesResilient object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSecretsPropertiesResilientWithDefaults() *SecretsPropertiesResilient {
	this := SecretsPropertiesResilient{}
	return &this
}

// GetApiKeyId returns the ApiKeyId field value
func (o *SecretsPropertiesResilient) GetApiKeyId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ApiKeyId
}

// GetApiKeyIdOk returns a tuple with the ApiKeyId field value
// and a boolean to check if the value has been set.
func (o *SecretsPropertiesResilient) GetApiKeyIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ApiKeyId, true
}

// SetApiKeyId sets field value
func (o *SecretsPropertiesResilient) SetApiKeyId(v string) {
	o.ApiKeyId = v
}

// GetApiKeySecret returns the ApiKeySecret field value
func (o *SecretsPropertiesResilient) GetApiKeySecret() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ApiKeySecret
}

// GetApiKeySecretOk returns a tuple with the ApiKeySecret field value
// and a boolean to check if the value has been set.
func (o *SecretsPropertiesResilient) GetApiKeySecretOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ApiKeySecret, true
}

// SetApiKeySecret sets field value
func (o *SecretsPropertiesResilient) SetApiKeySecret(v string) {
	o.ApiKeySecret = v
}

func (o SecretsPropertiesResilient) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SecretsPropertiesResilient) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["apiKeyId"] = o.ApiKeyId
	toSerialize["apiKeySecret"] = o.ApiKeySecret
	return toSerialize, nil
}

type NullableSecretsPropertiesResilient struct {
	value *SecretsPropertiesResilient
	isSet bool
}

func (v NullableSecretsPropertiesResilient) Get() *SecretsPropertiesResilient {
	return v.value
}

func (v *NullableSecretsPropertiesResilient) Set(val *SecretsPropertiesResilient) {
	v.value = val
	v.isSet = true
}

func (v NullableSecretsPropertiesResilient) IsSet() bool {
	return v.isSet
}

func (v *NullableSecretsPropertiesResilient) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSecretsPropertiesResilient(val *SecretsPropertiesResilient) *NullableSecretsPropertiesResilient {
	return &NullableSecretsPropertiesResilient{value: val, isSet: true}
}

func (v NullableSecretsPropertiesResilient) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSecretsPropertiesResilient) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
