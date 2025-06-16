# TimesliceMetricPercentileMetric

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the metric. Only valid options are A-Z | 
**Aggregation** | **string** | The aggregation type of the metric. Only valid option is \&quot;percentile\&quot; | 
**Field** | **string** | The field of the metric. | 
**Percentile** | **float32** | The percentile value. | 
**Filter** | Pointer to **string** | The filter to apply to the metric. | [optional] 

## Methods

### NewTimesliceMetricPercentileMetric

`func NewTimesliceMetricPercentileMetric(name string, aggregation string, field string, percentile float32, ) *TimesliceMetricPercentileMetric`

NewTimesliceMetricPercentileMetric instantiates a new TimesliceMetricPercentileMetric object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTimesliceMetricPercentileMetricWithDefaults

`func NewTimesliceMetricPercentileMetricWithDefaults() *TimesliceMetricPercentileMetric`

NewTimesliceMetricPercentileMetricWithDefaults instantiates a new TimesliceMetricPercentileMetric object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *TimesliceMetricPercentileMetric) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TimesliceMetricPercentileMetric) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TimesliceMetricPercentileMetric) SetName(v string)`

SetName sets Name field to given value.


### GetAggregation

`func (o *TimesliceMetricPercentileMetric) GetAggregation() string`

GetAggregation returns the Aggregation field if non-nil, zero value otherwise.

### GetAggregationOk

`func (o *TimesliceMetricPercentileMetric) GetAggregationOk() (*string, bool)`

GetAggregationOk returns a tuple with the Aggregation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregation

`func (o *TimesliceMetricPercentileMetric) SetAggregation(v string)`

SetAggregation sets Aggregation field to given value.


### GetField

`func (o *TimesliceMetricPercentileMetric) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *TimesliceMetricPercentileMetric) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *TimesliceMetricPercentileMetric) SetField(v string)`

SetField sets Field field to given value.


### GetPercentile

`func (o *TimesliceMetricPercentileMetric) GetPercentile() float32`

GetPercentile returns the Percentile field if non-nil, zero value otherwise.

### GetPercentileOk

`func (o *TimesliceMetricPercentileMetric) GetPercentileOk() (*float32, bool)`

GetPercentileOk returns a tuple with the Percentile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPercentile

`func (o *TimesliceMetricPercentileMetric) SetPercentile(v float32)`

SetPercentile sets Percentile field to given value.


### GetFilter

`func (o *TimesliceMetricPercentileMetric) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *TimesliceMetricPercentileMetric) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *TimesliceMetricPercentileMetric) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *TimesliceMetricPercentileMetric) HasFilter() bool`

HasFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


