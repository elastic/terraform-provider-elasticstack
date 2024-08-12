# ParamsPropertyApmAnomaly

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceName** | Pointer to **string** | The service name from APM | [optional] 
**TransactionType** | Pointer to **string** | The transaction type from APM | [optional] 
**WindowSize** | **float32** | The window size | 
**WindowUnit** | **string** | The window size unit | 
**Environment** | **string** | The environment from APM | 
**AnomalySeverityType** | **string** | The anomaly threshold value | 

## Methods

### NewParamsPropertyApmAnomaly

`func NewParamsPropertyApmAnomaly(windowSize float32, windowUnit string, environment string, anomalySeverityType string, ) *ParamsPropertyApmAnomaly`

NewParamsPropertyApmAnomaly instantiates a new ParamsPropertyApmAnomaly object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyApmAnomalyWithDefaults

`func NewParamsPropertyApmAnomalyWithDefaults() *ParamsPropertyApmAnomaly`

NewParamsPropertyApmAnomalyWithDefaults instantiates a new ParamsPropertyApmAnomaly object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceName

`func (o *ParamsPropertyApmAnomaly) GetServiceName() string`

GetServiceName returns the ServiceName field if non-nil, zero value otherwise.

### GetServiceNameOk

`func (o *ParamsPropertyApmAnomaly) GetServiceNameOk() (*string, bool)`

GetServiceNameOk returns a tuple with the ServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceName

`func (o *ParamsPropertyApmAnomaly) SetServiceName(v string)`

SetServiceName sets ServiceName field to given value.

### HasServiceName

`func (o *ParamsPropertyApmAnomaly) HasServiceName() bool`

HasServiceName returns a boolean if a field has been set.

### GetTransactionType

`func (o *ParamsPropertyApmAnomaly) GetTransactionType() string`

GetTransactionType returns the TransactionType field if non-nil, zero value otherwise.

### GetTransactionTypeOk

`func (o *ParamsPropertyApmAnomaly) GetTransactionTypeOk() (*string, bool)`

GetTransactionTypeOk returns a tuple with the TransactionType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTransactionType

`func (o *ParamsPropertyApmAnomaly) SetTransactionType(v string)`

SetTransactionType sets TransactionType field to given value.

### HasTransactionType

`func (o *ParamsPropertyApmAnomaly) HasTransactionType() bool`

HasTransactionType returns a boolean if a field has been set.

### GetWindowSize

`func (o *ParamsPropertyApmAnomaly) GetWindowSize() float32`

GetWindowSize returns the WindowSize field if non-nil, zero value otherwise.

### GetWindowSizeOk

`func (o *ParamsPropertyApmAnomaly) GetWindowSizeOk() (*float32, bool)`

GetWindowSizeOk returns a tuple with the WindowSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowSize

`func (o *ParamsPropertyApmAnomaly) SetWindowSize(v float32)`

SetWindowSize sets WindowSize field to given value.


### GetWindowUnit

`func (o *ParamsPropertyApmAnomaly) GetWindowUnit() string`

GetWindowUnit returns the WindowUnit field if non-nil, zero value otherwise.

### GetWindowUnitOk

`func (o *ParamsPropertyApmAnomaly) GetWindowUnitOk() (*string, bool)`

GetWindowUnitOk returns a tuple with the WindowUnit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWindowUnit

`func (o *ParamsPropertyApmAnomaly) SetWindowUnit(v string)`

SetWindowUnit sets WindowUnit field to given value.


### GetEnvironment

`func (o *ParamsPropertyApmAnomaly) GetEnvironment() string`

GetEnvironment returns the Environment field if non-nil, zero value otherwise.

### GetEnvironmentOk

`func (o *ParamsPropertyApmAnomaly) GetEnvironmentOk() (*string, bool)`

GetEnvironmentOk returns a tuple with the Environment field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironment

`func (o *ParamsPropertyApmAnomaly) SetEnvironment(v string)`

SetEnvironment sets Environment field to given value.


### GetAnomalySeverityType

`func (o *ParamsPropertyApmAnomaly) GetAnomalySeverityType() string`

GetAnomalySeverityType returns the AnomalySeverityType field if non-nil, zero value otherwise.

### GetAnomalySeverityTypeOk

`func (o *ParamsPropertyApmAnomaly) GetAnomalySeverityTypeOk() (*string, bool)`

GetAnomalySeverityTypeOk returns a tuple with the AnomalySeverityType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnomalySeverityType

`func (o *ParamsPropertyApmAnomaly) SetAnomalySeverityType(v string)`

SetAnomalySeverityType sets AnomalySeverityType field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


