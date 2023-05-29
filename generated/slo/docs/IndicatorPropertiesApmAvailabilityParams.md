# IndicatorPropertiesApmAvailabilityParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Service** | **string** | The APM service name | 
**Environment** | **string** | The APM service environment or \&quot;*\&quot; | 
**TransactionType** | **string** | The APM transaction type or \&quot;*\&quot; | 
**TransactionName** | **string** | The APM transaction name or \&quot;*\&quot; | 
**Filter** | Pointer to **string** | KQL query used for filtering the data | [optional] 
**Index** | **string** | The index used by APM metrics | 

## Methods

### NewIndicatorPropertiesApmAvailabilityParams

`func NewIndicatorPropertiesApmAvailabilityParams(service string, environment string, transactionType string, transactionName string, index string, ) *IndicatorPropertiesApmAvailabilityParams`

NewIndicatorPropertiesApmAvailabilityParams instantiates a new IndicatorPropertiesApmAvailabilityParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesApmAvailabilityParamsWithDefaults

`func NewIndicatorPropertiesApmAvailabilityParamsWithDefaults() *IndicatorPropertiesApmAvailabilityParams`

NewIndicatorPropertiesApmAvailabilityParamsWithDefaults instantiates a new IndicatorPropertiesApmAvailabilityParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetService

`func (o *IndicatorPropertiesApmAvailabilityParams) GetService() string`

GetService returns the Service field if non-nil, zero value otherwise.

### GetServiceOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetServiceOk() (*string, bool)`

GetServiceOk returns a tuple with the Service field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetService

`func (o *IndicatorPropertiesApmAvailabilityParams) SetService(v string)`

SetService sets Service field to given value.


### GetEnvironment

`func (o *IndicatorPropertiesApmAvailabilityParams) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *IndicatorPropertiesApmAvailabilityParams) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetTransactionType

`func (o *IndicatorPropertiesApmAvailabilityParams) GetTransactionType() string`

GetTransactionType returns the TransactionType field if non-nil, zero value otherwise.

### GetTransactionTypeOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetTransactionTypeOk() (*string, bool)`

GetTransactionTypeOk returns a tuple with the TransactionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionType

`func (o *IndicatorPropertiesApmAvailabilityParams) SetTransactionType(v string)`

SetTransactionType sets TransactionType field to given value.


### GetTransactionName

`func (o *IndicatorPropertiesApmAvailabilityParams) GetTransactionName() string`

GetTransactionName returns the TransactionName field if non-nil, zero value otherwise.

### GetTransactionNameOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetTransactionNameOk() (*string, bool)`

GetTransactionNameOk returns a tuple with the TransactionName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionName

`func (o *IndicatorPropertiesApmAvailabilityParams) SetTransactionName(v string)`

SetTransactionName sets TransactionName field to given value.


### GetFilter

`func (o *IndicatorPropertiesApmAvailabilityParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesApmAvailabilityParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesApmAvailabilityParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetIndex

`func (o *IndicatorPropertiesApmAvailabilityParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesApmAvailabilityParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesApmAvailabilityParams) SetIndex(v string)`

SetIndex sets Index field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


