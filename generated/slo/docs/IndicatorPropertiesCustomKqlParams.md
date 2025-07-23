# IndicatorPropertiesCustomKqlParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Index** | **string** | The index or index pattern to use | 
**DataViewId** | Pointer to **string** | The kibana data view id to use, primarily used to include data view runtime mappings. Make sure to save SLO again if you add/update run time fields to the data view and if those fields are being used in slo queries. | [optional] 
**Filter** | Pointer to [**KqlWithFilters**](KqlWithFilters.md) |  | [optional] 
**Good** | [**KqlWithFiltersGood**](KqlWithFiltersGood.md) |  | 
**Total** | [**KqlWithFiltersTotal**](KqlWithFiltersTotal.md) |  | 
**TimestampField** | **string** | The timestamp field used in the source indice.  | 

## Methods

### NewIndicatorPropertiesCustomKqlParams

`func NewIndicatorPropertiesCustomKqlParams(index string, good KqlWithFiltersGood, total KqlWithFiltersTotal, timestampField string, ) *IndicatorPropertiesCustomKqlParams`

NewIndicatorPropertiesCustomKqlParams instantiates a new IndicatorPropertiesCustomKqlParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesCustomKqlParamsWithDefaults

`func NewIndicatorPropertiesCustomKqlParamsWithDefaults() *IndicatorPropertiesCustomKqlParams`

NewIndicatorPropertiesCustomKqlParamsWithDefaults instantiates a new IndicatorPropertiesCustomKqlParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIndex

`func (o *IndicatorPropertiesCustomKqlParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesCustomKqlParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesCustomKqlParams) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetDataViewId

`func (o *IndicatorPropertiesCustomKqlParams) GetDataViewId() string`

GetDataViewId returns the DataViewId field if non-nil, zero value otherwise.

### GetDataViewIdOk

`func (o *IndicatorPropertiesCustomKqlParams) GetDataViewIdOk() (*string, bool)`

GetDataViewIdOk returns a tuple with the DataViewId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataViewId

`func (o *IndicatorPropertiesCustomKqlParams) SetDataViewId(v string)`

SetDataViewId sets DataViewId field to given value.

### HasDataViewId

`func (o *IndicatorPropertiesCustomKqlParams) HasDataViewId() bool`

HasDataViewId returns a boolean if a field has been set.

### GetFilter

`func (o *IndicatorPropertiesCustomKqlParams) GetFilter() KqlWithFilters`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesCustomKqlParams) GetFilterOk() (*KqlWithFilters, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesCustomKqlParams) SetFilter(v KqlWithFilters)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesCustomKqlParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetGood

`func (o *IndicatorPropertiesCustomKqlParams) GetGood() KqlWithFiltersGood`

GetGood returns the Good field if non-nil, zero value otherwise.

### GetGoodOk

`func (o *IndicatorPropertiesCustomKqlParams) GetGoodOk() (*KqlWithFiltersGood, bool)`

GetGoodOk returns a tuple with the Good field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGood

`func (o *IndicatorPropertiesCustomKqlParams) SetGood(v KqlWithFiltersGood)`

SetGood sets Good field to given value.


### GetTotal

`func (o *IndicatorPropertiesCustomKqlParams) GetTotal() KqlWithFiltersTotal`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *IndicatorPropertiesCustomKqlParams) GetTotalOk() (*KqlWithFiltersTotal, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *IndicatorPropertiesCustomKqlParams) SetTotal(v KqlWithFiltersTotal)`

SetTotal sets Total field to given value.


### GetTimestampField

`func (o *IndicatorPropertiesCustomKqlParams) GetTimestampField() string`

GetTimestampField returns the TimestampField field if non-nil, zero value otherwise.

### GetTimestampFieldOk

`func (o *IndicatorPropertiesCustomKqlParams) GetTimestampFieldOk() (*string, bool)`

GetTimestampFieldOk returns a tuple with the TimestampField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestampField

`func (o *IndicatorPropertiesCustomKqlParams) SetTimestampField(v string)`

SetTimestampField sets TimestampField field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


