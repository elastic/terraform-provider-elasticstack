# IndicatorPropertiesTimesliceMetricParamsMetric

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Metrics** | [**[]IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner**](IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner.md) | List of metrics with their name, aggregation type, and field. | 
**Equation** | **string** | The equation to calculate the metric. | 
**Comparator** | **string** | The comparator to use to compare the equation to the threshold. | 
**Threshold** | **float32** | The threshold used to determine if the metric is a good slice or not. | 

## Methods

### NewIndicatorPropertiesTimesliceMetricParamsMetric

`func NewIndicatorPropertiesTimesliceMetricParamsMetric(metrics []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, equation string, comparator string, threshold float32, ) *IndicatorPropertiesTimesliceMetricParamsMetric`

NewIndicatorPropertiesTimesliceMetricParamsMetric instantiates a new IndicatorPropertiesTimesliceMetricParamsMetric object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesTimesliceMetricParamsMetricWithDefaults

`func NewIndicatorPropertiesTimesliceMetricParamsMetricWithDefaults() *IndicatorPropertiesTimesliceMetricParamsMetric`

NewIndicatorPropertiesTimesliceMetricParamsMetricWithDefaults instantiates a new IndicatorPropertiesTimesliceMetricParamsMetric object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMetrics

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetMetrics() []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner`

GetMetrics returns the Metrics field if non-nil, zero value otherwise.

### GetMetricsOk

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetMetricsOk() (*[]IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner, bool)`

GetMetricsOk returns a tuple with the Metrics field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetrics

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetMetrics(v []IndicatorPropertiesTimesliceMetricParamsMetricMetricsInner)`

SetMetrics sets Metrics field to given value.


### GetEquation

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetEquation() string`

GetEquation returns the Equation field if non-nil, zero value otherwise.

### GetEquationOk

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetEquationOk() (*string, bool)`

GetEquationOk returns a tuple with the Equation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEquation

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetEquation(v string)`

SetEquation sets Equation field to given value.


### GetComparator

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetComparator() string`

GetComparator returns the Comparator field if non-nil, zero value otherwise.

### GetComparatorOk

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetComparatorOk() (*string, bool)`

GetComparatorOk returns a tuple with the Comparator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetComparator

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetComparator(v string)`

SetComparator sets Comparator field to given value.


### GetThreshold

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetThreshold() float32`

GetThreshold returns the Threshold field if non-nil, zero value otherwise.

### GetThresholdOk

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) GetThresholdOk() (*float32, bool)`

GetThresholdOk returns a tuple with the Threshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreshold

`func (o *IndicatorPropertiesTimesliceMetricParamsMetric) SetThreshold(v float32)`

SetThreshold sets Threshold field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


