# TimesliceMetricDocCountMetric

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the metric. Only valid options are A-Z | 
**Aggregation** | **string** | The aggregation type of the metric. Only valid option is \&quot;doc_count\&quot; | 
**Filter** | Pointer to **string** | The filter to apply to the metric. | [optional] 

## Methods

### NewTimesliceMetricDocCountMetric

`func NewTimesliceMetricDocCountMetric(name string, aggregation string, ) *TimesliceMetricDocCountMetric`

NewTimesliceMetricDocCountMetric instantiates a new TimesliceMetricDocCountMetric object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTimesliceMetricDocCountMetricWithDefaults

`func NewTimesliceMetricDocCountMetricWithDefaults() *TimesliceMetricDocCountMetric`

NewTimesliceMetricDocCountMetricWithDefaults instantiates a new TimesliceMetricDocCountMetric object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *TimesliceMetricDocCountMetric) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TimesliceMetricDocCountMetric) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TimesliceMetricDocCountMetric) SetName(v string)`

SetName sets Name field to given value.


### GetAggregation

`func (o *TimesliceMetricDocCountMetric) GetAggregation() string`

GetAggregation returns the Aggregation field if non-nil, zero value otherwise.

### GetAggregationOk

`func (o *TimesliceMetricDocCountMetric) GetAggregationOk() (*string, bool)`

GetAggregationOk returns a tuple with the Aggregation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAggregation

`func (o *TimesliceMetricDocCountMetric) SetAggregation(v string)`

SetAggregation sets Aggregation field to given value.


### GetFilter

`func (o *TimesliceMetricDocCountMetric) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *TimesliceMetricDocCountMetric) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *TimesliceMetricDocCountMetric) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *TimesliceMetricDocCountMetric) HasFilter() bool`

HasFilter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


