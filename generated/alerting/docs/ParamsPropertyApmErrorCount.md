# ParamsPropertyApmErrorCount

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceName** | Pointer to **string** | The service name from APM | [optional] 
**WindowSize** | **float32** | The window size | 
**WindowUnit** | **string** | The window size unit | 
**Environment** | **string** | The environment from APM | 
**Threshold** | **float32** | The error count threshold value | 
**GroupBy** | Pointer to **[]string** |  | [optional] 
**ErrorGroupingKey** | Pointer to **string** |  | [optional] 

## Methods

### NewParamsPropertyApmErrorCount

`func NewParamsPropertyApmErrorCount(windowSize float32, windowUnit string, environment string, threshold float32, ) *ParamsPropertyApmErrorCount`

NewParamsPropertyApmErrorCount instantiates a new ParamsPropertyApmErrorCount object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyApmErrorCountWithDefaults

`func NewParamsPropertyApmErrorCountWithDefaults() *ParamsPropertyApmErrorCount`

NewParamsPropertyApmErrorCountWithDefaults instantiates a new ParamsPropertyApmErrorCount object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceName

`func (o *ParamsPropertyApmErrorCount) GetServiceName() string`

GetServiceName returns the ServiceName field if non-nil, zero value otherwise.

### GetServiceNameOk

`func (o *ParamsPropertyApmErrorCount) GetServiceNameOk() (*string, bool)`

GetServiceNameOk returns a tuple with the ServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceName

`func (o *ParamsPropertyApmErrorCount) SetServiceName(v string)`

SetServiceName sets ServiceName field to given value.

### HasServiceName

`func (o *ParamsPropertyApmErrorCount) HasServiceName() bool`

HasServiceName returns a boolean if a field has been set.

### GetWindowSize

`func (o *ParamsPropertyApmErrorCount) GetWindowSize() float32`

GetWindowSize returns the WindowSize field if non-nil, zero value otherwise.

### GetWindowSizeOk

`func (o *ParamsPropertyApmErrorCount) GetWindowSizeOk() (*float32, bool)`

GetWindowSizeOk returns a tuple with the WindowSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowSize

`func (o *ParamsPropertyApmErrorCount) SetWindowSize(v float32)`

SetWindowSize sets WindowSize field to given value.


### GetWindowUnit

`func (o *ParamsPropertyApmErrorCount) GetWindowUnit() string`

GetWindowUnit returns the WindowUnit field if non-nil, zero value otherwise.

### GetWindowUnitOk

`func (o *ParamsPropertyApmErrorCount) GetWindowUnitOk() (*string, bool)`

GetWindowUnitOk returns a tuple with the WindowUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowUnit

`func (o *ParamsPropertyApmErrorCount) SetWindowUnit(v string)`

SetWindowUnit sets WindowUnit field to given value.


### GetEnvironment

`func (o *ParamsPropertyApmErrorCount) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *ParamsPropertyApmErrorCount) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *ParamsPropertyApmErrorCount) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetThreshold

`func (o *ParamsPropertyApmErrorCount) GetThreshold() float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *ParamsPropertyApmErrorCount) GetThresholdOk() (*float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *ParamsPropertyApmErrorCount) SetThreshold(v float32)`

SetThreshold sets Threshold field to given value.


### GetGroupBy

`func (o *ParamsPropertyApmErrorCount) GetGroupBy() []string`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ParamsPropertyApmErrorCount) GetGroupByOk() (*[]string, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ParamsPropertyApmErrorCount) SetGroupBy(v []string)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ParamsPropertyApmErrorCount) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.

### GetErrorGroupingKey

`func (o *ParamsPropertyApmErrorCount) GetErrorGroupingKey() string`

GetErrorGroupingKey returns the ErrorGroupingKey field if non-nil, zero value otherwise.

### GetErrorGroupingKeyOk

`func (o *ParamsPropertyApmErrorCount) GetErrorGroupingKeyOk() (*string, bool)`

GetErrorGroupingKeyOk returns a tuple with the ErrorGroupingKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetErrorGroupingKey

`func (o *ParamsPropertyApmErrorCount) SetErrorGroupingKey(v string)`

SetErrorGroupingKey sets ErrorGroupingKey field to given value.

### HasErrorGroupingKey

`func (o *ParamsPropertyApmErrorCount) HasErrorGroupingKey() bool`

HasErrorGroupingKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


