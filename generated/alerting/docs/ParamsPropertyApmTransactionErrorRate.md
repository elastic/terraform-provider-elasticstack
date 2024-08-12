# ParamsPropertyApmTransactionErrorRate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceName** | Pointer to **string** | The service name from APM | [optional] 
**TransactionType** | Pointer to **string** | The transaction type from APM | [optional] 
**TransactionName** | Pointer to **string** | The transaction name from APM | [optional] 
**WindowSize** | **float32** | The window size | 
**WindowUnit** | **string** | The window size unit | 
**Environment** | **string** | The environment from APM | 
**Threshold** | **float32** | The error rate threshold value | 
**GroupBy** | Pointer to **[]string** |  | [optional] 

## Methods

### NewParamsPropertyApmTransactionErrorRate

`func NewParamsPropertyApmTransactionErrorRate(windowSize float32, windowUnit string, environment string, threshold float32, ) *ParamsPropertyApmTransactionErrorRate`

NewParamsPropertyApmTransactionErrorRate instantiates a new ParamsPropertyApmTransactionErrorRate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyApmTransactionErrorRateWithDefaults

`func NewParamsPropertyApmTransactionErrorRateWithDefaults() *ParamsPropertyApmTransactionErrorRate`

NewParamsPropertyApmTransactionErrorRateWithDefaults instantiates a new ParamsPropertyApmTransactionErrorRate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceName

`func (o *ParamsPropertyApmTransactionErrorRate) GetServiceName() string`

GetServiceName returns the ServiceName field if non-nil, zero value otherwise.

### GetServiceNameOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetServiceNameOk() (*string, bool)`

GetServiceNameOk returns a tuple with the ServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceName

`func (o *ParamsPropertyApmTransactionErrorRate) SetServiceName(v string)`

SetServiceName sets ServiceName field to given value.

### HasServiceName

`func (o *ParamsPropertyApmTransactionErrorRate) HasServiceName() bool`

HasServiceName returns a boolean if a field has been set.

### GetTransactionType

`func (o *ParamsPropertyApmTransactionErrorRate) GetTransactionType() string`

GetTransactionType returns the TransactionType field if non-nil, zero value otherwise.

### GetTransactionTypeOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetTransactionTypeOk() (*string, bool)`

GetTransactionTypeOk returns a tuple with the TransactionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionType

`func (o *ParamsPropertyApmTransactionErrorRate) SetTransactionType(v string)`

SetTransactionType sets TransactionType field to given value.

### HasTransactionType

`func (o *ParamsPropertyApmTransactionErrorRate) HasTransactionType() bool`

HasTransactionType returns a boolean if a field has been set.

### GetTransactionName

`func (o *ParamsPropertyApmTransactionErrorRate) GetTransactionName() string`

GetTransactionName returns the TransactionName field if non-nil, zero value otherwise.

### GetTransactionNameOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetTransactionNameOk() (*string, bool)`

GetTransactionNameOk returns a tuple with the TransactionName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionName

`func (o *ParamsPropertyApmTransactionErrorRate) SetTransactionName(v string)`

SetTransactionName sets TransactionName field to given value.

### HasTransactionName

`func (o *ParamsPropertyApmTransactionErrorRate) HasTransactionName() bool`

HasTransactionName returns a boolean if a field has been set.

### GetWindowSize

`func (o *ParamsPropertyApmTransactionErrorRate) GetWindowSize() float32`

GetWindowSize returns the WindowSize field if non-nil, zero value otherwise.

### GetWindowSizeOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetWindowSizeOk() (*float32, bool)`

GetWindowSizeOk returns a tuple with the WindowSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowSize

`func (o *ParamsPropertyApmTransactionErrorRate) SetWindowSize(v float32)`

SetWindowSize sets WindowSize field to given value.


### GetWindowUnit

`func (o *ParamsPropertyApmTransactionErrorRate) GetWindowUnit() string`

GetWindowUnit returns the WindowUnit field if non-nil, zero value otherwise.

### GetWindowUnitOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetWindowUnitOk() (*string, bool)`

GetWindowUnitOk returns a tuple with the WindowUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowUnit

`func (o *ParamsPropertyApmTransactionErrorRate) SetWindowUnit(v string)`

SetWindowUnit sets WindowUnit field to given value.


### GetEnvironment

`func (o *ParamsPropertyApmTransactionErrorRate) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *ParamsPropertyApmTransactionErrorRate) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetThreshold

`func (o *ParamsPropertyApmTransactionErrorRate) GetThreshold() float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetThresholdOk() (*float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *ParamsPropertyApmTransactionErrorRate) SetThreshold(v float32)`

SetThreshold sets Threshold field to given value.


### GetGroupBy

`func (o *ParamsPropertyApmTransactionErrorRate) GetGroupBy() []string`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ParamsPropertyApmTransactionErrorRate) GetGroupByOk() (*[]string, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ParamsPropertyApmTransactionErrorRate) SetGroupBy(v []string)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ParamsPropertyApmTransactionErrorRate) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


