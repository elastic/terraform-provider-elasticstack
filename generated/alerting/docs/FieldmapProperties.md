# FieldmapProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Array** | Pointer to **bool** | Indicates whether the field is an array. | [optional] 
**Dynamic** | Pointer to **bool** | Indicates whether it is a dynamic field mapping. | [optional] 
**Format** | Pointer to **string** | Indicates the format of the field. For example, if the &#x60;type&#x60; is &#x60;date_range&#x60;, the &#x60;format&#x60; can be &#x60;epoch_millis||strict_date_optional_time&#x60;.  | [optional] 
**IgnoreAbove** | Pointer to **int32** | Specifies the maximum length of a string field. Longer strings are not indexed or stored. | [optional] 
**Index** | Pointer to **bool** | Indicates whether field values are indexed. | [optional] 
**Path** | Pointer to **string** | TBD | [optional] 
**Properties** | Pointer to  | Details about the object properties. This property is applicable when &#x60;type&#x60; is &#x60;object&#x60;.  | [optional] 
**Required** | Pointer to **bool** | Indicates whether the field is required. | [optional] 
**ScalingFactor** | Pointer to **int32** | The scaling factor to use when encoding values. This property is applicable when &#x60;type&#x60; is &#x60;scaled_float&#x60;. Values will be multiplied by this factor at index time and rounded to the closest long value.   | [optional] 
**Type** | Pointer to **string** | Specifies the data type for the field. | [optional] 

## Methods

### NewFieldmapProperties

`func NewFieldmapProperties() *FieldmapProperties`

NewFieldmapProperties instantiates a new FieldmapProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldmapPropertiesWithDefaults

`func NewFieldmapPropertiesWithDefaults() *FieldmapProperties`

NewFieldmapPropertiesWithDefaults instantiates a new FieldmapProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetArray

`func (o *FieldmapProperties) GetArray() bool`

GetArray returns the Array field if non-nil, zero value otherwise.

### GetArrayOk

`func (o *FieldmapProperties) GetArrayOk() (*bool, bool)`

GetArrayOk returns a tuple with the Array field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArray

`func (o *FieldmapProperties) SetArray(v bool)`

SetArray sets Array field to given value.

### HasArray

`func (o *FieldmapProperties) HasArray() bool`

HasArray returns a boolean if a field has been set.

### GetDynamic

`func (o *FieldmapProperties) GetDynamic() bool`

GetDynamic returns the Dynamic field if non-nil, zero value otherwise.

### GetDynamicOk

`func (o *FieldmapProperties) GetDynamicOk() (*bool, bool)`

GetDynamicOk returns a tuple with the Dynamic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDynamic

`func (o *FieldmapProperties) SetDynamic(v bool)`

SetDynamic sets Dynamic field to given value.

### HasDynamic

`func (o *FieldmapProperties) HasDynamic() bool`

HasDynamic returns a boolean if a field has been set.

### GetFormat

`func (o *FieldmapProperties) GetFormat() string`

GetFormat returns the Format field if non-nil, zero value otherwise.

### GetFormatOk

`func (o *FieldmapProperties) GetFormatOk() (*string, bool)`

GetFormatOk returns a tuple with the Format field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFormat

`func (o *FieldmapProperties) SetFormat(v string)`

SetFormat sets Format field to given value.

### HasFormat

`func (o *FieldmapProperties) HasFormat() bool`

HasFormat returns a boolean if a field has been set.

### GetIgnoreAbove

`func (o *FieldmapProperties) GetIgnoreAbove() int32`

GetIgnoreAbove returns the IgnoreAbove field if non-nil, zero value otherwise.

### GetIgnoreAboveOk

`func (o *FieldmapProperties) GetIgnoreAboveOk() (*int32, bool)`

GetIgnoreAboveOk returns a tuple with the IgnoreAbove field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIgnoreAbove

`func (o *FieldmapProperties) SetIgnoreAbove(v int32)`

SetIgnoreAbove sets IgnoreAbove field to given value.

### HasIgnoreAbove

`func (o *FieldmapProperties) HasIgnoreAbove() bool`

HasIgnoreAbove returns a boolean if a field has been set.

### GetIndex

`func (o *FieldmapProperties) GetIndex() bool`

GetIndex returns the Index field if non-nil, zero value otherwise.

### GetIndexOk

`func (o *FieldmapProperties) GetIndexOk() (*bool, bool)`

GetIndexOk returns a tuple with the Index field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndex

`func (o *FieldmapProperties) SetIndex(v bool)`

SetIndex sets Index field to given value.

### HasIndex

`func (o *FieldmapProperties) HasIndex() bool`

HasIndex returns a boolean if a field has been set.

### GetPath

`func (o *FieldmapProperties) GetPath() string`

GetPath returns the Path field if non-nil, zero value otherwise.

### GetPathOk

`func (o *FieldmapProperties) GetPathOk() (*string, bool)`

GetPathOk returns a tuple with the Path field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPath

`func (o *FieldmapProperties) SetPath(v string)`

SetPath sets Path field to given value.

### HasPath

`func (o *FieldmapProperties) HasPath() bool`

HasPath returns a boolean if a field has been set.

### GetProperties

`func (o *FieldmapProperties) GetProperties() map[string]FieldmapPropertiesPropertiesValue`

GetProperties returns the Properties field if non-nil, zero value otherwise.

### GetPropertiesOk

`func (o *FieldmapProperties) GetPropertiesOk() (*map[string]FieldmapPropertiesPropertiesValue, bool)`

GetPropertiesOk returns a tuple with the Properties field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProperties

`func (o *FieldmapProperties) SetProperties(v map[string]FieldmapPropertiesPropertiesValue)`

SetProperties sets Properties field to given value.

### HasProperties

`func (o *FieldmapProperties) HasProperties() bool`

HasProperties returns a boolean if a field has been set.

### SetPropertiesNil

`func (o *FieldmapProperties) SetPropertiesNil(b bool)`

 SetPropertiesNil sets the value for Properties to be an explicit nil

### UnsetProperties
`func (o *FieldmapProperties) UnsetProperties()`

UnsetProperties ensures that no value is present for Properties, not even an explicit nil
### GetRequired

`func (o *FieldmapProperties) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *FieldmapProperties) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *FieldmapProperties) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *FieldmapProperties) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### GetScalingFactor

`func (o *FieldmapProperties) GetScalingFactor() int32`

GetScalingFactor returns the ScalingFactor field if non-nil, zero value otherwise.

### GetScalingFactorOk

`func (o *FieldmapProperties) GetScalingFactorOk() (*int32, bool)`

GetScalingFactorOk returns a tuple with the ScalingFactor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScalingFactor

`func (o *FieldmapProperties) SetScalingFactor(v int32)`

SetScalingFactor sets ScalingFactor field to given value.

### HasScalingFactor

`func (o *FieldmapProperties) HasScalingFactor() bool`

HasScalingFactor returns a boolean if a field has been set.

### GetType

`func (o *FieldmapProperties) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *FieldmapProperties) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *FieldmapProperties) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *FieldmapProperties) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


