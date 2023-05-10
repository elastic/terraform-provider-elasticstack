# IndicatorPropertiesCustomKqlParams

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Index** | **string** | The index or index pattern to use | 
**Filter** | Pointer to **string** | the KQL query to filter the documents with. | [optional] 
**Good** | Pointer to **string** | the KQL query used to define the good events. | [optional] 
**Total** | Pointer to **string** | the KQL query used to define all events. | [optional] 
**TimestampField** | **string** | The timestamp field used in the source indice. If not specified, @timestamp will be used.  | 

## Methods

### NewIndicatorPropertiesCustomKqlParams

`func NewIndicatorPropertiesCustomKqlParams(index string, timestampField string, ) *IndicatorPropertiesCustomKqlParams`

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


### GetFilter

`func (o *IndicatorPropertiesCustomKqlParams) GetFilter() string`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *IndicatorPropertiesCustomKqlParams) GetFilterOk() (*string, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *IndicatorPropertiesCustomKqlParams) SetFilter(v string)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *IndicatorPropertiesCustomKqlParams) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetGood

`func (o *IndicatorPropertiesCustomKqlParams) GetGood() string`

GetGood returns the Good field if non-nil, zero value otherwise.

### GetGoodOk

`func (o *IndicatorPropertiesCustomKqlParams) GetGoodOk() (*string, bool)`

GetGoodOk returns a tuple with the Good field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGood

`func (o *IndicatorPropertiesCustomKqlParams) SetGood(v string)`

SetGood sets Good field to given value.

### HasGood

`func (o *IndicatorPropertiesCustomKqlParams) HasGood() bool`

HasGood returns a boolean if a field has been set.

### GetTotal

`func (o *IndicatorPropertiesCustomKqlParams) GetTotal() string`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *IndicatorPropertiesCustomKqlParams) GetTotalOk() (*string, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *IndicatorPropertiesCustomKqlParams) SetTotal(v string)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *IndicatorPropertiesCustomKqlParams) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

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


