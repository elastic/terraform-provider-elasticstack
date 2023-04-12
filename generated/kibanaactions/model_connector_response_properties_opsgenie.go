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

// checks if the ConnectorResponsePropertiesOpsgenie type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ConnectorResponsePropertiesOpsgenie{}

// ConnectorResponsePropertiesOpsgenie struct for ConnectorResponsePropertiesOpsgenie
type ConnectorResponsePropertiesOpsgenie struct {
	Config ConfigPropertiesOpsgenie `json:"config"`
	// The type of connector.
	ConnectorTypeId string `json:"connector_type_id"`
	// The identifier for the connector.
	Id string `json:"id"`
	// Indicates whether the connector type is deprecated.
	IsDeprecated bool `json:"is_deprecated"`
	// Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type.
	IsMissingSecrets *bool `json:"is_missing_secrets,omitempty"`
	// Indicates whether it is a preconfigured connector. If true, the `config` and `is_missing_secrets` properties are omitted from the response.
	IsPreconfigured bool `json:"is_preconfigured"`
	// The display name for the connector.
	Name string `json:"name"`
}

// NewConnectorResponsePropertiesOpsgenie instantiates a new ConnectorResponsePropertiesOpsgenie object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorResponsePropertiesOpsgenie(config ConfigPropertiesOpsgenie, connectorTypeId string, id string, isDeprecated bool, isPreconfigured bool, name string) *ConnectorResponsePropertiesOpsgenie {
	this := ConnectorResponsePropertiesOpsgenie{}
	this.Config = config
	this.ConnectorTypeId = connectorTypeId
	this.Id = id
	this.IsDeprecated = isDeprecated
	this.IsPreconfigured = isPreconfigured
	this.Name = name
	return &this
}

// NewConnectorResponsePropertiesOpsgenieWithDefaults instantiates a new ConnectorResponsePropertiesOpsgenie object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorResponsePropertiesOpsgenieWithDefaults() *ConnectorResponsePropertiesOpsgenie {
	this := ConnectorResponsePropertiesOpsgenie{}
	return &this
}

// GetConfig returns the Config field value
func (o *ConnectorResponsePropertiesOpsgenie) GetConfig() ConfigPropertiesOpsgenie {
	if o == nil {
		var ret ConfigPropertiesOpsgenie
		return ret
	}

	return o.Config
}

// GetConfigOk returns a tuple with the Config field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetConfigOk() (*ConfigPropertiesOpsgenie, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Config, true
}

// SetConfig sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetConfig(v ConfigPropertiesOpsgenie) {
	o.Config = v
}

// GetConnectorTypeId returns the ConnectorTypeId field value
func (o *ConnectorResponsePropertiesOpsgenie) GetConnectorTypeId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ConnectorTypeId
}

// GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetConnectorTypeIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ConnectorTypeId, true
}

// SetConnectorTypeId sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetConnectorTypeId(v string) {
	o.ConnectorTypeId = v
}

// GetId returns the Id field value
func (o *ConnectorResponsePropertiesOpsgenie) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetId(v string) {
	o.Id = v
}

// GetIsDeprecated returns the IsDeprecated field value
func (o *ConnectorResponsePropertiesOpsgenie) GetIsDeprecated() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.IsDeprecated
}

// GetIsDeprecatedOk returns a tuple with the IsDeprecated field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetIsDeprecatedOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.IsDeprecated, true
}

// SetIsDeprecated sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetIsDeprecated(v bool) {
	o.IsDeprecated = v
}

// GetIsMissingSecrets returns the IsMissingSecrets field value if set, zero value otherwise.
func (o *ConnectorResponsePropertiesOpsgenie) GetIsMissingSecrets() bool {
	if o == nil || IsNil(o.IsMissingSecrets) {
		var ret bool
		return ret
	}
	return *o.IsMissingSecrets
}

// GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetIsMissingSecretsOk() (*bool, bool) {
	if o == nil || IsNil(o.IsMissingSecrets) {
		return nil, false
	}
	return o.IsMissingSecrets, true
}

// HasIsMissingSecrets returns a boolean if a field has been set.
func (o *ConnectorResponsePropertiesOpsgenie) HasIsMissingSecrets() bool {
	if o != nil && !IsNil(o.IsMissingSecrets) {
		return true
	}

	return false
}

// SetIsMissingSecrets gets a reference to the given bool and assigns it to the IsMissingSecrets field.
func (o *ConnectorResponsePropertiesOpsgenie) SetIsMissingSecrets(v bool) {
	o.IsMissingSecrets = &v
}

// GetIsPreconfigured returns the IsPreconfigured field value
func (o *ConnectorResponsePropertiesOpsgenie) GetIsPreconfigured() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.IsPreconfigured
}

// GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetIsPreconfiguredOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.IsPreconfigured, true
}

// SetIsPreconfigured sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetIsPreconfigured(v bool) {
	o.IsPreconfigured = v
}

// GetName returns the Name field value
func (o *ConnectorResponsePropertiesOpsgenie) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ConnectorResponsePropertiesOpsgenie) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ConnectorResponsePropertiesOpsgenie) SetName(v string) {
	o.Name = v
}

func (o ConnectorResponsePropertiesOpsgenie) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ConnectorResponsePropertiesOpsgenie) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["config"] = o.Config
	toSerialize["connector_type_id"] = o.ConnectorTypeId
	toSerialize["id"] = o.Id
	toSerialize["is_deprecated"] = o.IsDeprecated
	if !IsNil(o.IsMissingSecrets) {
		toSerialize["is_missing_secrets"] = o.IsMissingSecrets
	}
	toSerialize["is_preconfigured"] = o.IsPreconfigured
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

type NullableConnectorResponsePropertiesOpsgenie struct {
	value *ConnectorResponsePropertiesOpsgenie
	isSet bool
}

func (v NullableConnectorResponsePropertiesOpsgenie) Get() *ConnectorResponsePropertiesOpsgenie {
	return v.value
}

func (v *NullableConnectorResponsePropertiesOpsgenie) Set(val *ConnectorResponsePropertiesOpsgenie) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorResponsePropertiesOpsgenie) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorResponsePropertiesOpsgenie) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorResponsePropertiesOpsgenie(val *ConnectorResponsePropertiesOpsgenie) *NullableConnectorResponsePropertiesOpsgenie {
	return &NullableConnectorResponsePropertiesOpsgenie{value: val, isSet: true}
}

func (v NullableConnectorResponsePropertiesOpsgenie) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorResponsePropertiesOpsgenie) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}