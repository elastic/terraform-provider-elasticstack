# IndicatorPropertiesCustomMetricParamsGood

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metrics** | [**[]IndicatorPropertiesCustomMetricParamsGoodMetricsInner**](IndicatorPropertiesCustomMetricParamsGoodMetricsInner.md) | List of metrics with their name, aggregation type, and field. | 
**Equation** | **string** | The equation to calculate the \&quot;good\&quot; metric. | 

## Methods

### NewIndicatorPropertiesCustomMetricParamsGood

`func NewIndicatorPropertiesCustomMetricParamsGood(metrics []IndicatorPropertiesCustomMetricParamsGoodMetricsInner, equation string, ) *IndicatorPropertiesCustomMetricParamsGood`

NewIndicatorPropertiesCustomMetricParamsGood instantiates a new IndicatorPropertiesCustomMetricParamsGood object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesCustomMetricParamsGoodWithDefaults

`func NewIndicatorPropertiesCustomMetricParamsGoodWithDefaults() *IndicatorPropertiesCustomMetricParamsGood`

NewIndicatorPropertiesCustomMetricParamsGoodWithDefaults instantiates a new IndicatorPropertiesCustomMetricParamsGood object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetrics

`func (o *IndicatorPropertiesCustomMetricParamsGood) GetMetrics() []IndicatorPropertiesCustomMetricParamsGoodMetricsInner`

GetMetrics returns the Metrics field if non-nil, zero value otherwise.

### GetMetricsOk

`func (o *IndicatorPropertiesCustomMetricParamsGood) GetMetricsOk() (*[]IndicatorPropertiesCustomMetricParamsGoodMetricsInner, bool)`

GetMetricsOk returns a tuple with the Metrics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetrics

`func (o *IndicatorPropertiesCustomMetricParamsGood) SetMetrics(v []IndicatorPropertiesCustomMetricParamsGoodMetricsInner)`

SetMetrics sets Metrics field to given value.


### GetEquation

`func (o *IndicatorPropertiesCustomMetricParamsGood) GetEquation() string`

GetEquation returns the Equation field if non-nil, zero value otherwise.

### GetEquationOk

`func (o *IndicatorPropertiesCustomMetricParamsGood) GetEquationOk() (*string, bool)`

GetEquationOk returns a tuple with the Equation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEquation

`func (o *IndicatorPropertiesCustomMetricParamsGood) SetEquation(v string)`

SetEquation sets Equation field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


