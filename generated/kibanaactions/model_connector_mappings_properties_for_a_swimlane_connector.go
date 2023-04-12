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

// checks if the ConnectorMappingsPropertiesForASwimlaneConnector type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ConnectorMappingsPropertiesForASwimlaneConnector{}

// ConnectorMappingsPropertiesForASwimlaneConnector The field mapping.
type ConnectorMappingsPropertiesForASwimlaneConnector struct {
	AlertIdConfig     *AlertIdentifierMapping `json:"alertIdConfig,omitempty"`
	CaseIdConfig      *CaseIdentifierMapping  `json:"caseIdConfig,omitempty"`
	CaseNameConfig    *CaseNameMapping        `json:"caseNameConfig,omitempty"`
	CommentsConfig    *CaseCommentMapping     `json:"commentsConfig,omitempty"`
	DescriptionConfig *CaseDescriptionMapping `json:"descriptionConfig,omitempty"`
	RuleNameConfig    *RuleNameMapping        `json:"ruleNameConfig,omitempty"`
	SeverityConfig    *SeverityMapping        `json:"severityConfig,omitempty"`
}

// NewConnectorMappingsPropertiesForASwimlaneConnector instantiates a new ConnectorMappingsPropertiesForASwimlaneConnector object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConnectorMappingsPropertiesForASwimlaneConnector() *ConnectorMappingsPropertiesForASwimlaneConnector {
	this := ConnectorMappingsPropertiesForASwimlaneConnector{}
	return &this
}

// NewConnectorMappingsPropertiesForASwimlaneConnectorWithDefaults instantiates a new ConnectorMappingsPropertiesForASwimlaneConnector object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConnectorMappingsPropertiesForASwimlaneConnectorWithDefaults() *ConnectorMappingsPropertiesForASwimlaneConnector {
	this := ConnectorMappingsPropertiesForASwimlaneConnector{}
	return &this
}

// GetAlertIdConfig returns the AlertIdConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetAlertIdConfig() AlertIdentifierMapping {
	if o == nil || IsNil(o.AlertIdConfig) {
		var ret AlertIdentifierMapping
		return ret
	}
	return *o.AlertIdConfig
}

// GetAlertIdConfigOk returns a tuple with the AlertIdConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetAlertIdConfigOk() (*AlertIdentifierMapping, bool) {
	if o == nil || IsNil(o.AlertIdConfig) {
		return nil, false
	}
	return o.AlertIdConfig, true
}

// HasAlertIdConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasAlertIdConfig() bool {
	if o != nil && !IsNil(o.AlertIdConfig) {
		return true
	}

	return false
}

// SetAlertIdConfig gets a reference to the given AlertIdentifierMapping and assigns it to the AlertIdConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetAlertIdConfig(v AlertIdentifierMapping) {
	o.AlertIdConfig = &v
}

// GetCaseIdConfig returns the CaseIdConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCaseIdConfig() CaseIdentifierMapping {
	if o == nil || IsNil(o.CaseIdConfig) {
		var ret CaseIdentifierMapping
		return ret
	}
	return *o.CaseIdConfig
}

// GetCaseIdConfigOk returns a tuple with the CaseIdConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCaseIdConfigOk() (*CaseIdentifierMapping, bool) {
	if o == nil || IsNil(o.CaseIdConfig) {
		return nil, false
	}
	return o.CaseIdConfig, true
}

// HasCaseIdConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasCaseIdConfig() bool {
	if o != nil && !IsNil(o.CaseIdConfig) {
		return true
	}

	return false
}

// SetCaseIdConfig gets a reference to the given CaseIdentifierMapping and assigns it to the CaseIdConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetCaseIdConfig(v CaseIdentifierMapping) {
	o.CaseIdConfig = &v
}

// GetCaseNameConfig returns the CaseNameConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCaseNameConfig() CaseNameMapping {
	if o == nil || IsNil(o.CaseNameConfig) {
		var ret CaseNameMapping
		return ret
	}
	return *o.CaseNameConfig
}

// GetCaseNameConfigOk returns a tuple with the CaseNameConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCaseNameConfigOk() (*CaseNameMapping, bool) {
	if o == nil || IsNil(o.CaseNameConfig) {
		return nil, false
	}
	return o.CaseNameConfig, true
}

// HasCaseNameConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasCaseNameConfig() bool {
	if o != nil && !IsNil(o.CaseNameConfig) {
		return true
	}

	return false
}

// SetCaseNameConfig gets a reference to the given CaseNameMapping and assigns it to the CaseNameConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetCaseNameConfig(v CaseNameMapping) {
	o.CaseNameConfig = &v
}

// GetCommentsConfig returns the CommentsConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCommentsConfig() CaseCommentMapping {
	if o == nil || IsNil(o.CommentsConfig) {
		var ret CaseCommentMapping
		return ret
	}
	return *o.CommentsConfig
}

// GetCommentsConfigOk returns a tuple with the CommentsConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetCommentsConfigOk() (*CaseCommentMapping, bool) {
	if o == nil || IsNil(o.CommentsConfig) {
		return nil, false
	}
	return o.CommentsConfig, true
}

// HasCommentsConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasCommentsConfig() bool {
	if o != nil && !IsNil(o.CommentsConfig) {
		return true
	}

	return false
}

// SetCommentsConfig gets a reference to the given CaseCommentMapping and assigns it to the CommentsConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetCommentsConfig(v CaseCommentMapping) {
	o.CommentsConfig = &v
}

// GetDescriptionConfig returns the DescriptionConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetDescriptionConfig() CaseDescriptionMapping {
	if o == nil || IsNil(o.DescriptionConfig) {
		var ret CaseDescriptionMapping
		return ret
	}
	return *o.DescriptionConfig
}

// GetDescriptionConfigOk returns a tuple with the DescriptionConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetDescriptionConfigOk() (*CaseDescriptionMapping, bool) {
	if o == nil || IsNil(o.DescriptionConfig) {
		return nil, false
	}
	return o.DescriptionConfig, true
}

// HasDescriptionConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasDescriptionConfig() bool {
	if o != nil && !IsNil(o.DescriptionConfig) {
		return true
	}

	return false
}

// SetDescriptionConfig gets a reference to the given CaseDescriptionMapping and assigns it to the DescriptionConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetDescriptionConfig(v CaseDescriptionMapping) {
	o.DescriptionConfig = &v
}

// GetRuleNameConfig returns the RuleNameConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetRuleNameConfig() RuleNameMapping {
	if o == nil || IsNil(o.RuleNameConfig) {
		var ret RuleNameMapping
		return ret
	}
	return *o.RuleNameConfig
}

// GetRuleNameConfigOk returns a tuple with the RuleNameConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetRuleNameConfigOk() (*RuleNameMapping, bool) {
	if o == nil || IsNil(o.RuleNameConfig) {
		return nil, false
	}
	return o.RuleNameConfig, true
}

// HasRuleNameConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasRuleNameConfig() bool {
	if o != nil && !IsNil(o.RuleNameConfig) {
		return true
	}

	return false
}

// SetRuleNameConfig gets a reference to the given RuleNameMapping and assigns it to the RuleNameConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetRuleNameConfig(v RuleNameMapping) {
	o.RuleNameConfig = &v
}

// GetSeverityConfig returns the SeverityConfig field value if set, zero value otherwise.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetSeverityConfig() SeverityMapping {
	if o == nil || IsNil(o.SeverityConfig) {
		var ret SeverityMapping
		return ret
	}
	return *o.SeverityConfig
}

// GetSeverityConfigOk returns a tuple with the SeverityConfig field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) GetSeverityConfigOk() (*SeverityMapping, bool) {
	if o == nil || IsNil(o.SeverityConfig) {
		return nil, false
	}
	return o.SeverityConfig, true
}

// HasSeverityConfig returns a boolean if a field has been set.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) HasSeverityConfig() bool {
	if o != nil && !IsNil(o.SeverityConfig) {
		return true
	}

	return false
}

// SetSeverityConfig gets a reference to the given SeverityMapping and assigns it to the SeverityConfig field.
func (o *ConnectorMappingsPropertiesForASwimlaneConnector) SetSeverityConfig(v SeverityMapping) {
	o.SeverityConfig = &v
}

func (o ConnectorMappingsPropertiesForASwimlaneConnector) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ConnectorMappingsPropertiesForASwimlaneConnector) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.AlertIdConfig) {
		toSerialize["alertIdConfig"] = o.AlertIdConfig
	}
	if !IsNil(o.CaseIdConfig) {
		toSerialize["caseIdConfig"] = o.CaseIdConfig
	}
	if !IsNil(o.CaseNameConfig) {
		toSerialize["caseNameConfig"] = o.CaseNameConfig
	}
	if !IsNil(o.CommentsConfig) {
		toSerialize["commentsConfig"] = o.CommentsConfig
	}
	if !IsNil(o.DescriptionConfig) {
		toSerialize["descriptionConfig"] = o.DescriptionConfig
	}
	if !IsNil(o.RuleNameConfig) {
		toSerialize["ruleNameConfig"] = o.RuleNameConfig
	}
	if !IsNil(o.SeverityConfig) {
		toSerialize["severityConfig"] = o.SeverityConfig
	}
	return toSerialize, nil
}

type NullableConnectorMappingsPropertiesForASwimlaneConnector struct {
	value *ConnectorMappingsPropertiesForASwimlaneConnector
	isSet bool
}

func (v NullableConnectorMappingsPropertiesForASwimlaneConnector) Get() *ConnectorMappingsPropertiesForASwimlaneConnector {
	return v.value
}

func (v *NullableConnectorMappingsPropertiesForASwimlaneConnector) Set(val *ConnectorMappingsPropertiesForASwimlaneConnector) {
	v.value = val
	v.isSet = true
}

func (v NullableConnectorMappingsPropertiesForASwimlaneConnector) IsSet() bool {
	return v.isSet
}

func (v *NullableConnectorMappingsPropertiesForASwimlaneConnector) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConnectorMappingsPropertiesForASwimlaneConnector(val *ConnectorMappingsPropertiesForASwimlaneConnector) *NullableConnectorMappingsPropertiesForASwimlaneConnector {
	return &NullableConnectorMappingsPropertiesForASwimlaneConnector{value: val, isSet: true}
}

func (v NullableConnectorMappingsPropertiesForASwimlaneConnector) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConnectorMappingsPropertiesForASwimlaneConnector) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}