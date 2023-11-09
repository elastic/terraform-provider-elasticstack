# IndicatorPropertiesCustomMetricParamsTotalMetricsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the metric. Only valid options are A-Z | 
**Aggregation** | **string** | The aggregation type of the metric. Only valid option is \&quot;sum\&quot; | 
**Field** | **string** | The field of the metric. | 
**Filter** | Pointer to **string** | The filter to apply to the metric. | [optional] 

## Methods

### NewIndicatorPropertiesCustomMetricParamsTotalMetricsInner

`func NewIndicatorPropertiesCustomMetricParamsTotalMetricsInner(name string, aggregation string, field string, ) *IndicatorPropertiesCustomMetricParamsTotalMetricsInner`

NewIndicatorPropertiesCustomMetricParamsTotalMetricsInner instantiates a new IndicatorPropertiesCustomMetricParamsTotalMetricsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesCustomMetricParamsTotalMetricsInnerWithDefaults

`func NewIndicatorPropertiesCustomMetricParamsTotalMetricsInnerWithDefaults() *IndicatorPropertiesCustomMetricParamsTotalMetricsInner`

NewIndicatorPropertiesCustomMetricParamsTotalMetricsInnerWithDefaults instantiates a new IndicatorPropertiesCustomMetricParamsTotalMetricsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) SetName(v string)`

SetName sets Name field to given value.


### GetAggregation

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetAggregation() string`

GetAggregation returns the Aggregation field if non-nil, zero value otherwise.

### GetAggregationOk

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetAggregationOk() (*string, bool)`

GetAggregationOk returns a tuple with the Aggregation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregation

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) SetAggregation(v string)`

SetAggregation sets Aggregation field to given value.


### GetField

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) SetField(v string)`

SetField sets Field field to given value.


### GetFilter

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesCustomMetricParamsTotalMetricsInner) HasFilter() bool`

HasFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


