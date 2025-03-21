/*
SLOs

OpenAPI schema for SLOs endpoints

API version: 1.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package slo

import (
	"encoding/json"
)

// checks if the IndicatorPropertiesApmLatencyParams type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &IndicatorPropertiesApmLatencyParams{}

// IndicatorPropertiesApmLatencyParams An object containing the indicator parameters.
type IndicatorPropertiesApmLatencyParams struct {
	// The APM service name
	Service string `json:"service"`
	// The APM service environment or \"*\"
	Environment string `json:"environment"`
	// The APM transaction type or \"*\"
	TransactionType string `json:"transactionType"`
	// The APM transaction name or \"*\"
	TransactionName string `json:"transactionName"`
	// KQL query used for filtering the data
	Filter *string `json:"filter,omitempty"`
	// The index used by APM metrics
	Index string `json:"index"`
	// The latency threshold in milliseconds
	Threshold float64 `json:"threshold"`
}

// NewIndicatorPropertiesApmLatencyParams instantiates a new IndicatorPropertiesApmLatencyParams object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIndicatorPropertiesApmLatencyParams(service string, environment string, transactionType string, transactionName string, index string, threshold float64) *IndicatorPropertiesApmLatencyParams {
	this := IndicatorPropertiesApmLatencyParams{}
	this.Service = service
	this.Environment = environment
	this.TransactionType = transactionType
	this.TransactionName = transactionName
	this.Index = index
	this.Threshold = threshold
	return &this
}

// NewIndicatorPropertiesApmLatencyParamsWithDefaults instantiates a new IndicatorPropertiesApmLatencyParams object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIndicatorPropertiesApmLatencyParamsWithDefaults() *IndicatorPropertiesApmLatencyParams {
	this := IndicatorPropertiesApmLatencyParams{}
	return &this
}

// GetService returns the Service field value
func (o *IndicatorPropertiesApmLatencyParams) GetService() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Service
}

// GetServiceOk returns a tuple with the Service field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetServiceOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Service, true
}

// SetService sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetService(v string) {
	o.Service = v
}

// GetEnvironment returns the Environment field value
func (o *IndicatorPropertiesApmLatencyParams) GetEnvironment() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Environment
}

// GetEnvironmentOk returns a tuple with the Environment field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetEnvironmentOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Environment, true
}

// SetEnvironment sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetEnvironment(v string) {
	o.Environment = v
}

// GetTransactionType returns the TransactionType field value
func (o *IndicatorPropertiesApmLatencyParams) GetTransactionType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TransactionType
}

// GetTransactionTypeOk returns a tuple with the TransactionType field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetTransactionTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TransactionType, true
}

// SetTransactionType sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetTransactionType(v string) {
	o.TransactionType = v
}

// GetTransactionName returns the TransactionName field value
func (o *IndicatorPropertiesApmLatencyParams) GetTransactionName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TransactionName
}

// GetTransactionNameOk returns a tuple with the TransactionName field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetTransactionNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TransactionName, true
}

// SetTransactionName sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetTransactionName(v string) {
	o.TransactionName = v
}

// GetFilter returns the Filter field value if set, zero value otherwise.
func (o *IndicatorPropertiesApmLatencyParams) GetFilter() string {
	if o == nil || IsNil(o.Filter) {
		var ret string
		return ret
	}
	return *o.Filter
}

// GetFilterOk returns a tuple with the Filter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetFilterOk() (*string, bool) {
	if o == nil || IsNil(o.Filter) {
		return nil, false
	}
	return o.Filter, true
}

// HasFilter returns a boolean if a field has been set.
func (o *IndicatorPropertiesApmLatencyParams) HasFilter() bool {
	if o != nil && !IsNil(o.Filter) {
		return true
	}

	return false
}

// SetFilter gets a reference to the given string and assigns it to the Filter field.
func (o *IndicatorPropertiesApmLatencyParams) SetFilter(v string) {
	o.Filter = &v
}

// GetIndex returns the Index field value
func (o *IndicatorPropertiesApmLatencyParams) GetIndex() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Index
}

// GetIndexOk returns a tuple with the Index field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetIndexOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Index, true
}

// SetIndex sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetIndex(v string) {
	o.Index = v
}

// GetThreshold returns the Threshold field value
func (o *IndicatorPropertiesApmLatencyParams) GetThreshold() float64 {
	if o == nil {
		var ret float64
		return ret
	}

	return o.Threshold
}

// GetThresholdOk returns a tuple with the Threshold field value
// and a boolean to check if the value has been set.
func (o *IndicatorPropertiesApmLatencyParams) GetThresholdOk() (*float64, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Threshold, true
}

// SetThreshold sets field value
func (o *IndicatorPropertiesApmLatencyParams) SetThreshold(v float64) {
	o.Threshold = v
}

func (o IndicatorPropertiesApmLatencyParams) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o IndicatorPropertiesApmLatencyParams) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["service"] = o.Service
	toSerialize["environment"] = o.Environment
	toSerialize["transactionType"] = o.TransactionType
	toSerialize["transactionName"] = o.TransactionName
	if !IsNil(o.Filter) {
		toSerialize["filter"] = o.Filter
	}
	toSerialize["index"] = o.Index
	toSerialize["threshold"] = o.Threshold
	return toSerialize, nil
}

type NullableIndicatorPropertiesApmLatencyParams struct {
	value *IndicatorPropertiesApmLatencyParams
	isSet bool
}

func (v NullableIndicatorPropertiesApmLatencyParams) Get() *IndicatorPropertiesApmLatencyParams {
	return v.value
}

func (v *NullableIndicatorPropertiesApmLatencyParams) Set(val *IndicatorPropertiesApmLatencyParams) {
	v.value = val
	v.isSet = true
}

func (v NullableIndicatorPropertiesApmLatencyParams) IsSet() bool {
	return v.isSet
}

func (v *NullableIndicatorPropertiesApmLatencyParams) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIndicatorPropertiesApmLatencyParams(val *IndicatorPropertiesApmLatencyParams) *NullableIndicatorPropertiesApmLatencyParams {
	return &NullableIndicatorPropertiesApmLatencyParams{value: val, isSet: true}
}

func (v NullableIndicatorPropertiesApmLatencyParams) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIndicatorPropertiesApmLatencyParams) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
