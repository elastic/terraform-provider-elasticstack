# IndicatorPropertiesCustomMetricParamsTotal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metrics** | [**[]IndicatorPropertiesCustomMetricParamsGoodMetricsInner**](IndicatorPropertiesCustomMetricParamsGoodMetricsInner.md) | List of metrics with their name, aggregation type, and field. | 
**Equation** | **string** | The equation to calculate the \&quot;total\&quot; metric. | 

## Methods

### NewIndicatorPropertiesCustomMetricParamsTotal

`func NewIndicatorPropertiesCustomMetricParamsTotal(metrics []IndicatorPropertiesCustomMetricParamsGoodMetricsInner, equation string, ) *IndicatorPropertiesCustomMetricParamsTotal`

NewIndicatorPropertiesCustomMetricParamsTotal instantiates a new IndicatorPropertiesCustomMetricParamsTotal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesCustomMetricParamsTotalWithDefaults

`func NewIndicatorPropertiesCustomMetricParamsTotalWithDefaults() *IndicatorPropertiesCustomMetricParamsTotal`

NewIndicatorPropertiesCustomMetricParamsTotalWithDefaults instantiates a new IndicatorPropertiesCustomMetricParamsTotal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetrics

`func (o *IndicatorPropertiesCustomMetricParamsTotal) GetMetrics() []IndicatorPropertiesCustomMetricParamsGoodMetricsInner`

GetMetrics returns the Metrics field if non-nil, zero value otherwise.

### GetMetricsOk

`func (o *IndicatorPropertiesCustomMetricParamsTotal) GetMetricsOk() (*[]IndicatorPropertiesCustomMetricParamsGoodMetricsInner, bool)`

GetMetricsOk returns a tuple with the Metrics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetrics

`func (o *IndicatorPropertiesCustomMetricParamsTotal) SetMetrics(v []IndicatorPropertiesCustomMetricParamsGoodMetricsInner)`

SetMetrics sets Metrics field to given value.


### GetEquation

`func (o *IndicatorPropertiesCustomMetricParamsTotal) GetEquation() string`

GetEquation returns the Equation field if non-nil, zero value otherwise.

### GetEquationOk

`func (o *IndicatorPropertiesCustomMetricParamsTotal) GetEquationOk() (*string, bool)`

GetEquationOk returns a tuple with the Equation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEquation

`func (o *IndicatorPropertiesCustomMetricParamsTotal) SetEquation(v string)`

SetEquation sets Equation field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


