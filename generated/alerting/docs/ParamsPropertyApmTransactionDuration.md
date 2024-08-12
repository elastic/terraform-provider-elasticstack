# ParamsPropertyApmTransactionDuration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceName** | Pointer to **string** | The service name from APM | [optional] 
**TransactionType** | Pointer to **string** | The transaction type from APM | [optional] 
**TransactionName** | Pointer to **string** | The transaction name from APM | [optional] 
**WindowSize** | **float32** | The window size | 
**WindowUnit** | **string** | ç | 
**Environment** | **string** |  | 
**Threshold** | **float32** | The latency threshold value | 
**GroupBy** | Pointer to **[]string** |  | [optional] 
**AggregationType** | **string** |  | 

## Methods

### NewParamsPropertyApmTransactionDuration

`func NewParamsPropertyApmTransactionDuration(windowSize float32, windowUnit string, environment string, threshold float32, aggregationType string, ) *ParamsPropertyApmTransactionDuration`

NewParamsPropertyApmTransactionDuration instantiates a new ParamsPropertyApmTransactionDuration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyApmTransactionDurationWithDefaults

`func NewParamsPropertyApmTransactionDurationWithDefaults() *ParamsPropertyApmTransactionDuration`

NewParamsPropertyApmTransactionDurationWithDefaults instantiates a new ParamsPropertyApmTransactionDuration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceName

`func (o *ParamsPropertyApmTransactionDuration) GetServiceName() string`

GetServiceName returns the ServiceName field if non-nil, zero value otherwise.

### GetServiceNameOk

`func (o *ParamsPropertyApmTransactionDuration) GetServiceNameOk() (*string, bool)`

GetServiceNameOk returns a tuple with the ServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceName

`func (o *ParamsPropertyApmTransactionDuration) SetServiceName(v string)`

SetServiceName sets ServiceName field to given value.

### HasServiceName

`func (o *ParamsPropertyApmTransactionDuration) HasServiceName() bool`

HasServiceName returns a boolean if a field has been set.

### GetTransactionType

`func (o *ParamsPropertyApmTransactionDuration) GetTransactionType() string`

GetTransactionType returns the TransactionType field if non-nil, zero value otherwise.

### GetTransactionTypeOk

`func (o *ParamsPropertyApmTransactionDuration) GetTransactionTypeOk() (*string, bool)`

GetTransactionTypeOk returns a tuple with the TransactionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionType

`func (o *ParamsPropertyApmTransactionDuration) SetTransactionType(v string)`

SetTransactionType sets TransactionType field to given value.

### HasTransactionType

`func (o *ParamsPropertyApmTransactionDuration) HasTransactionType() bool`

HasTransactionType returns a boolean if a field has been set.

### GetTransactionName

`func (o *ParamsPropertyApmTransactionDuration) GetTransactionName() string`

GetTransactionName returns the TransactionName field if non-nil, zero value otherwise.

### GetTransactionNameOk

`func (o *ParamsPropertyApmTransactionDuration) GetTransactionNameOk() (*string, bool)`

GetTransactionNameOk returns a tuple with the TransactionName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionName

`func (o *ParamsPropertyApmTransactionDuration) SetTransactionName(v string)`

SetTransactionName sets TransactionName field to given value.

### HasTransactionName

`func (o *ParamsPropertyApmTransactionDuration) HasTransactionName() bool`

HasTransactionName returns a boolean if a field has been set.

### GetWindowSize

`func (o *ParamsPropertyApmTransactionDuration) GetWindowSize() float32`

GetWindowSize returns the WindowSize field if non-nil, zero value otherwise.

### GetWindowSizeOk

`func (o *ParamsPropertyApmTransactionDuration) GetWindowSizeOk() (*float32, bool)`

GetWindowSizeOk returns a tuple with the WindowSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowSize

`func (o *ParamsPropertyApmTransactionDuration) SetWindowSize(v float32)`

SetWindowSize sets WindowSize field to given value.


### GetWindowUnit

`func (o *ParamsPropertyApmTransactionDuration) GetWindowUnit() string`

GetWindowUnit returns the WindowUnit field if non-nil, zero value otherwise.

### GetWindowUnitOk

`func (o *ParamsPropertyApmTransactionDuration) GetWindowUnitOk() (*string, bool)`

GetWindowUnitOk returns a tuple with the WindowUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowUnit

`func (o *ParamsPropertyApmTransactionDuration) SetWindowUnit(v string)`

SetWindowUnit sets WindowUnit field to given value.


### GetEnvironment

`func (o *ParamsPropertyApmTransactionDuration) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *ParamsPropertyApmTransactionDuration) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *ParamsPropertyApmTransactionDuration) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetThreshold

`func (o *ParamsPropertyApmTransactionDuration) GetThreshold() float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *ParamsPropertyApmTransactionDuration) GetThresholdOk() (*float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *ParamsPropertyApmTransactionDuration) SetThreshold(v float32)`

SetThreshold sets Threshold field to given value.


### GetGroupBy

`func (o *ParamsPropertyApmTransactionDuration) GetGroupBy() []string`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ParamsPropertyApmTransactionDuration) GetGroupByOk() (*[]string, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ParamsPropertyApmTransactionDuration) SetGroupBy(v []string)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ParamsPropertyApmTransactionDuration) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.

### GetAggregationType

`func (o *ParamsPropertyApmTransactionDuration) GetAggregationType() string`

GetAggregationType returns the AggregationType field if non-nil, zero value otherwise.

### GetAggregationTypeOk

`func (o *ParamsPropertyApmTransactionDuration) GetAggregationTypeOk() (*string, bool)`

GetAggregationTypeOk returns a tuple with the AggregationType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregationType

`func (o *ParamsPropertyApmTransactionDuration) SetAggregationType(v string)`

SetAggregationType sets AggregationType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


