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

// checks if the CreateConnectorRequestEmail type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CreateConnectorRequestEmail{}

// CreateConnectorRequestEmail The email connector uses the SMTP protocol to send mail messages, using an integration of Nodemailer. An exception is Microsoft Exchange, which uses HTTP protocol for sending emails, Send mail. Email message text is sent as both plain text and html text.
type CreateConnectorRequestEmail struct {
	// Defines properties for connectors when type is `.email`.
	Config map[string]interface{} `json:"config"`
	// The type of connector.
	ConnectorTypeId string `json:"connector_type_id"`
	// The display name for the connector.
	Name string `json:"name"`
	// Defines secrets for connectors when type is `.email`.
	Secrets map[string]interface{} `json:"secrets"`
}

// NewCreateConnectorRequestEmail instantiates a new CreateConnectorRequestEmail object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCreateConnectorRequestEmail(config map[string]interface{}, connectorTypeId string, name string, secrets map[string]interface{}) *CreateConnectorRequestEmail {
	this := CreateConnectorRequestEmail{}
	this.Config = config
	this.ConnectorTypeId = connectorTypeId
	this.Name = name
	this.Secrets = secrets
	return &this
}

// NewCreateConnectorRequestEmailWithDefaults instantiates a new CreateConnectorRequestEmail object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCreateConnectorRequestEmailWithDefaults() *CreateConnectorRequestEmail {
	this := CreateConnectorRequestEmail{}
	return &this
}

// GetConfig returns the Config field value
func (o *CreateConnectorRequestEmail) GetConfig() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Config
}

// GetConfigOk returns a tuple with the Config field value
// and a boolean to check if the value has been set.
func (o *CreateConnectorRequestEmail) GetConfigOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Config, true
}

// SetConfig sets field value
func (o *CreateConnectorRequestEmail) SetConfig(v map[string]interface{}) {
	o.Config = v
}

// GetConnectorTypeId returns the ConnectorTypeId field value
func (o *CreateConnectorRequestEmail) GetConnectorTypeId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ConnectorTypeId
}

// GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field value
// and a boolean to check if the value has been set.
func (o *CreateConnectorRequestEmail) GetConnectorTypeIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ConnectorTypeId, true
}

// SetConnectorTypeId sets field value
func (o *CreateConnectorRequestEmail) SetConnectorTypeId(v string) {
	o.ConnectorTypeId = v
}

// GetName returns the Name field value
func (o *CreateConnectorRequestEmail) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *CreateConnectorRequestEmail) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *CreateConnectorRequestEmail) SetName(v string) {
	o.Name = v
}

// GetSecrets returns the Secrets field value
func (o *CreateConnectorRequestEmail) GetSecrets() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Secrets
}

// GetSecretsOk returns a tuple with the Secrets field value
// and a boolean to check if the value has been set.
func (o *CreateConnectorRequestEmail) GetSecretsOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Secrets, true
}

// SetSecrets sets field value
func (o *CreateConnectorRequestEmail) SetSecrets(v map[string]interface{}) {
	o.Secrets = v
}

func (o CreateConnectorRequestEmail) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CreateConnectorRequestEmail) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["config"] = o.Config
	toSerialize["connector_type_id"] = o.ConnectorTypeId
	toSerialize["name"] = o.Name
	toSerialize["secrets"] = o.Secrets
	return toSerialize, nil
}

type NullableCreateConnectorRequestEmail struct {
	value *CreateConnectorRequestEmail
	isSet bool
}

func (v NullableCreateConnectorRequestEmail) Get() *CreateConnectorRequestEmail {
	return v.value
}

func (v *NullableCreateConnectorRequestEmail) Set(val *CreateConnectorRequestEmail) {
	v.value = val
	v.isSet = true
}

func (v NullableCreateConnectorRequestEmail) IsSet() bool {
	return v.isSet
}

func (v *NullableCreateConnectorRequestEmail) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCreateConnectorRequestEmail(val *CreateConnectorRequestEmail) *NullableCreateConnectorRequestEmail {
	return &NullableCreateConnectorRequestEmail{value: val, isSet: true}
}

func (v NullableCreateConnectorRequestEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCreateConnectorRequestEmail) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
