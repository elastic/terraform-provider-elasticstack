# IndicatorPropertiesTimesliceMetricParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Index** | **string** | The index or index pattern to use | 
**DataViewId** | Pointer to **string** | The kibana data view id to use, primarily used to include data view runtime mappings. Make sure to save SLO again if you add/update run time fields to the data view and if those fields are being used in slo queries. | [optional] 
**Filter** | Pointer to **string** | the KQL query to filter the documents with. | [optional] 
**TimestampField** | **string** | The timestamp field used in the source indice.  | 
**Metric** | [**IndicatorPropertiesTimesliceMetricParamsMetric**](IndicatorPropertiesTimesliceMetricParamsMetric.md) |  | 

## Methods

### NewIndicatorPropertiesTimesliceMetricParams

`func NewIndicatorPropertiesTimesliceMetricParams(index string, timestampField string, metric IndicatorPropertiesTimesliceMetricParamsMetric, ) *IndicatorPropertiesTimesliceMetricParams`

NewIndicatorPropertiesTimesliceMetricParams instantiates a new IndicatorPropertiesTimesliceMetricParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesTimesliceMetricParamsWithDefaults

`func NewIndicatorPropertiesTimesliceMetricParamsWithDefaults() *IndicatorPropertiesTimesliceMetricParams`

NewIndicatorPropertiesTimesliceMetricParamsWithDefaults instantiates a new IndicatorPropertiesTimesliceMetricParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIndex

`func (o *IndicatorPropertiesTimesliceMetricParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesTimesliceMetricParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesTimesliceMetricParams) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetDataViewId

`func (o *IndicatorPropertiesTimesliceMetricParams) GetDataViewId() string`

GetDataViewId returns the DataViewId field if non-nil, zero value otherwise.

### GetDataViewIdOk

`func (o *IndicatorPropertiesTimesliceMetricParams) GetDataViewIdOk() (*string, bool)`

GetDataViewIdOk returns a tuple with the DataViewId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataViewId

`func (o *IndicatorPropertiesTimesliceMetricParams) SetDataViewId(v string)`

SetDataViewId sets DataViewId field to given value.

### HasDataViewId

`func (o *IndicatorPropertiesTimesliceMetricParams) HasDataViewId() bool`

HasDataViewId returns a boolean if a field has been set.

### GetFilter

`func (o *IndicatorPropertiesTimesliceMetricParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesTimesliceMetricParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesTimesliceMetricParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesTimesliceMetricParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetTimestampField

`func (o *IndicatorPropertiesTimesliceMetricParams) GetTimestampField() string`

GetTimestampField returns the TimestampField field if non-nil, zero value otherwise.

### GetTimestampFieldOk

`func (o *IndicatorPropertiesTimesliceMetricParams) GetTimestampFieldOk() (*string, bool)`

GetTimestampFieldOk returns a tuple with the TimestampField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestampField

`func (o *IndicatorPropertiesTimesliceMetricParams) SetTimestampField(v string)`

SetTimestampField sets TimestampField field to given value.


### GetMetric

`func (o *IndicatorPropertiesTimesliceMetricParams) GetMetric() IndicatorPropertiesTimesliceMetricParamsMetric`

GetMetric returns the Metric field if non-nil, zero value otherwise.

### GetMetricOk

`func (o *IndicatorPropertiesTimesliceMetricParams) GetMetricOk() (*IndicatorPropertiesTimesliceMetricParamsMetric, bool)`

GetMetricOk returns a tuple with the Metric field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMetric

`func (o *IndicatorPropertiesTimesliceMetricParams) SetMetric(v IndicatorPropertiesTimesliceMetricParamsMetric)`

SetMetric sets Metric field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


