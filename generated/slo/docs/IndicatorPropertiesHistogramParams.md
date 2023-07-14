# IndicatorPropertiesHistogramParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Index** | **string** | The index or index pattern to use | 
**Filter** | Pointer to **string** | the KQL query to filter the documents with. | [optional] 
**TimestampField** | **string** | The timestamp field used in the source indice.  | 
**Good** | [**IndicatorPropertiesHistogramParamsGood**](IndicatorPropertiesHistogramParamsGood.md) |  | 
**Total** | [**IndicatorPropertiesHistogramParamsTotal**](IndicatorPropertiesHistogramParamsTotal.md) |  | 

## Methods

### NewIndicatorPropertiesHistogramParams

`func NewIndicatorPropertiesHistogramParams(index string, timestampField string, good IndicatorPropertiesHistogramParamsGood, total IndicatorPropertiesHistogramParamsTotal, ) *IndicatorPropertiesHistogramParams`

NewIndicatorPropertiesHistogramParams instantiates a new IndicatorPropertiesHistogramParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesHistogramParamsWithDefaults

`func NewIndicatorPropertiesHistogramParamsWithDefaults() *IndicatorPropertiesHistogramParams`

NewIndicatorPropertiesHistogramParamsWithDefaults instantiates a new IndicatorPropertiesHistogramParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIndex

`func (o *IndicatorPropertiesHistogramParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesHistogramParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesHistogramParams) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetFilter

`func (o *IndicatorPropertiesHistogramParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesHistogramParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesHistogramParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesHistogramParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetTimestampField

`func (o *IndicatorPropertiesHistogramParams) GetTimestampField() string`

GetTimestampField returns the TimestampField field if non-nil, zero value otherwise.

### GetTimestampFieldOk

`func (o *IndicatorPropertiesHistogramParams) GetTimestampFieldOk() (*string, bool)`

GetTimestampFieldOk returns a tuple with the TimestampField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestampField

`func (o *IndicatorPropertiesHistogramParams) SetTimestampField(v string)`

SetTimestampField sets TimestampField field to given value.


### GetGood

`func (o *IndicatorPropertiesHistogramParams) GetGood() IndicatorPropertiesHistogramParamsGood`

GetGood returns the Good field if non-nil, zero value otherwise.

### GetGoodOk

`func (o *IndicatorPropertiesHistogramParams) GetGoodOk() (*IndicatorPropertiesHistogramParamsGood, bool)`

GetGoodOk returns a tuple with the Good field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGood

`func (o *IndicatorPropertiesHistogramParams) SetGood(v IndicatorPropertiesHistogramParamsGood)`

SetGood sets Good field to given value.


### GetTotal

`func (o *IndicatorPropertiesHistogramParams) GetTotal() IndicatorPropertiesHistogramParamsTotal`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *IndicatorPropertiesHistogramParams) GetTotalOk() (*IndicatorPropertiesHistogramParamsTotal, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *IndicatorPropertiesHistogramParams) SetTotal(v IndicatorPropertiesHistogramParamsTotal)`

SetTotal sets Total field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


