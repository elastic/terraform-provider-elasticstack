# IndicatorPropertiesApmLatencyParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Service** | **string** | The APM service name | 
**Environment** | **string** | The APM service environment or \&quot;*\&quot; | 
**TransactionType** | **string** | The APM transaction type or \&quot;*\&quot; | 
**TransactionName** | **string** | The APM transaction name or \&quot;*\&quot; | 
**Filter** | Pointer to **string** | KQL query used for filtering the data | [optional] 
**Index** | **string** | The index used by APM metrics | 
**Threshold** | **float32** | The latency threshold in milliseconds | 

## Methods

### NewIndicatorPropertiesApmLatencyParams

`func NewIndicatorPropertiesApmLatencyParams(service string, environment string, transactionType string, transactionName string, index string, threshold float32, ) *IndicatorPropertiesApmLatencyParams`

NewIndicatorPropertiesApmLatencyParams instantiates a new IndicatorPropertiesApmLatencyParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesApmLatencyParamsWithDefaults

`func NewIndicatorPropertiesApmLatencyParamsWithDefaults() *IndicatorPropertiesApmLatencyParams`

NewIndicatorPropertiesApmLatencyParamsWithDefaults instantiates a new IndicatorPropertiesApmLatencyParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetService

`func (o *IndicatorPropertiesApmLatencyParams) GetService() string`

GetService returns the Service field if non-nil, zero value otherwise.

### GetServiceOk

`func (o *IndicatorPropertiesApmLatencyParams) GetServiceOk() (*string, bool)`

GetServiceOk returns a tuple with the Service field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetService

`func (o *IndicatorPropertiesApmLatencyParams) SetService(v string)`

SetService sets Service field to given value.


### GetEnvironment

`func (o *IndicatorPropertiesApmLatencyParams) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *IndicatorPropertiesApmLatencyParams) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *IndicatorPropertiesApmLatencyParams) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetTransactionType

`func (o *IndicatorPropertiesApmLatencyParams) GetTransactionType() string`

GetTransactionType returns the TransactionType field if non-nil, zero value otherwise.

### GetTransactionTypeOk

`func (o *IndicatorPropertiesApmLatencyParams) GetTransactionTypeOk() (*string, bool)`

GetTransactionTypeOk returns a tuple with the TransactionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionType

`func (o *IndicatorPropertiesApmLatencyParams) SetTransactionType(v string)`

SetTransactionType sets TransactionType field to given value.


### GetTransactionName

`func (o *IndicatorPropertiesApmLatencyParams) GetTransactionName() string`

GetTransactionName returns the TransactionName field if non-nil, zero value otherwise.

### GetTransactionNameOk

`func (o *IndicatorPropertiesApmLatencyParams) GetTransactionNameOk() (*string, bool)`

GetTransactionNameOk returns a tuple with the TransactionName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionName

`func (o *IndicatorPropertiesApmLatencyParams) SetTransactionName(v string)`

SetTransactionName sets TransactionName field to given value.


### GetFilter

`func (o *IndicatorPropertiesApmLatencyParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesApmLatencyParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesApmLatencyParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesApmLatencyParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetIndex

`func (o *IndicatorPropertiesApmLatencyParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesApmLatencyParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesApmLatencyParams) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetThreshold

`func (o *IndicatorPropertiesApmLatencyParams) GetThreshold() float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *IndicatorPropertiesApmLatencyParams) GetThresholdOk() (*float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *IndicatorPropertiesApmLatencyParams) SetThreshold(v float32)`

SetThreshold sets Threshold field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


