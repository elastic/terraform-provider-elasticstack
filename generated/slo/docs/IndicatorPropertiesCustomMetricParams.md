# IndicatorPropertiesCustomMetricParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Index** | **string** | The index or index pattern to use | 
**Filter** | Pointer to **string** | the KQL query to filter the documents with. | [optional] 
**TimestampField** | **string** | The timestamp field used in the source indice.  | 
**Good** | [**IndicatorPropertiesCustomMetricParamsGood**](IndicatorPropertiesCustomMetricParamsGood.md) |  | 
**Total** | [**IndicatorPropertiesCustomMetricParamsTotal**](IndicatorPropertiesCustomMetricParamsTotal.md) |  | 

## Methods

### NewIndicatorPropertiesCustomMetricParams

`func NewIndicatorPropertiesCustomMetricParams(index string, timestampField string, good IndicatorPropertiesCustomMetricParamsGood, total IndicatorPropertiesCustomMetricParamsTotal, ) *IndicatorPropertiesCustomMetricParams`

NewIndicatorPropertiesCustomMetricParams instantiates a new IndicatorPropertiesCustomMetricParams object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIndicatorPropertiesCustomMetricParamsWithDefaults

`func NewIndicatorPropertiesCustomMetricParamsWithDefaults() *IndicatorPropertiesCustomMetricParams`

NewIndicatorPropertiesCustomMetricParamsWithDefaults instantiates a new IndicatorPropertiesCustomMetricParams object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIndex

`func (o *IndicatorPropertiesCustomMetricParams) GetIndex() string`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *IndicatorPropertiesCustomMetricParams) GetIndexOk() (*string, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *IndicatorPropertiesCustomMetricParams) SetIndex(v string)`

SetIndex sets Index field to given value.


### GetFilter

`func (o *IndicatorPropertiesCustomMetricParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesCustomMetricParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesCustomMetricParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesCustomMetricParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetTimestampField

`func (o *IndicatorPropertiesCustomMetricParams) GetTimestampField() string`

GetTimestampField returns the TimestampField field if non-nil, zero value otherwise.

### GetTimestampFieldOk

`func (o *IndicatorPropertiesCustomMetricParams) GetTimestampFieldOk() (*string, bool)`

GetTimestampFieldOk returns a tuple with the TimestampField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestampField

`func (o *IndicatorPropertiesCustomMetricParams) SetTimestampField(v string)`

SetTimestampField sets TimestampField field to given value.


### GetGood

`func (o *IndicatorPropertiesCustomMetricParams) GetGood() IndicatorPropertiesCustomMetricParamsGood`

GetGood returns the Good field if non-nil, zero value otherwise.

### GetGoodOk

`func (o *IndicatorPropertiesCustomMetricParams) GetGoodOk() (*IndicatorPropertiesCustomMetricParamsGood, bool)`

GetGoodOk returns a tuple with the Good field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGood

`func (o *IndicatorPropertiesCustomMetricParams) SetGood(v IndicatorPropertiesCustomMetricParamsGood)`

SetGood sets Good field to given value.


### GetTotal

`func (o *IndicatorPropertiesCustomMetricParams) GetTotal() IndicatorPropertiesCustomMetricParamsTotal`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *IndicatorPropertiesCustomMetricParams) GetTotalOk() (*IndicatorPropertiesCustomMetricParamsTotal, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *IndicatorPropertiesCustomMetricParams) SetTotal(v IndicatorPropertiesCustomMetricParamsTotal)`

SetTotal sets Total field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


