# UpdateDataViewRequestObjectDataView

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowNoIndex** | Pointer to **interface{}** | Allows the data view saved object to exist before the data is available. | [optional] 
**FieldFormats** | Pointer to **interface{}** | A map of field formats by field name. | [optional] 
**Fields** | Pointer to **interface{}** |  | [optional] 
**Name** | Pointer to **interface{}** |  | [optional] 
**RuntimeFieldMap** | Pointer to **interface{}** | A map of runtime field definitions by field name. | [optional] 
**SourceFilters** | Pointer to **interface{}** | The array of field names you want to filter out in Discover. | [optional] 
**TimeFieldName** | Pointer to **interface{}** | The timestamp field name, which you use for time-based data views. | [optional] 
**Title** | Pointer to **interface{}** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | [optional] 
**Type** | Pointer to **interface{}** | When set to &#x60;rollup&#x60;, identifies the rollup data views. | [optional] 
**TypeMeta** | Pointer to **interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 

## Methods

### NewUpdateDataViewRequestObjectDataView

`func NewUpdateDataViewRequestObjectDataView() *UpdateDataViewRequestObjectDataView`

NewUpdateDataViewRequestObjectDataView instantiates a new UpdateDataViewRequestObjectDataView object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateDataViewRequestObjectDataViewWithDefaults

`func NewUpdateDataViewRequestObjectDataViewWithDefaults() *UpdateDataViewRequestObjectDataView`

NewUpdateDataViewRequestObjectDataViewWithDefaults instantiates a new UpdateDataViewRequestObjectDataView object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowNoIndex

`func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndex() interface{}`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndexOk() (*interface{}, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *UpdateDataViewRequestObjectDataView) SetAllowNoIndex(v interface{})`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *UpdateDataViewRequestObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### SetAllowNoIndexNil

`func (o *UpdateDataViewRequestObjectDataView) SetAllowNoIndexNil(b bool)`

 SetAllowNoIndexNil sets the value for AllowNoIndex to be an explicit nil

### UnsetAllowNoIndex
`func (o *UpdateDataViewRequestObjectDataView) UnsetAllowNoIndex()`

UnsetAllowNoIndex ensures that no value is present for AllowNoIndex, not even an explicit nil
### GetFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) GetFieldFormats() interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *UpdateDataViewRequestObjectDataView) GetFieldFormatsOk() (*interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) SetFieldFormats(v interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### SetFieldFormatsNil

`func (o *UpdateDataViewRequestObjectDataView) SetFieldFormatsNil(b bool)`

 SetFieldFormatsNil sets the value for FieldFormats to be an explicit nil

### UnsetFieldFormats
`func (o *UpdateDataViewRequestObjectDataView) UnsetFieldFormats()`

UnsetFieldFormats ensures that no value is present for FieldFormats, not even an explicit nil
### GetFields

`func (o *UpdateDataViewRequestObjectDataView) GetFields() interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *UpdateDataViewRequestObjectDataView) GetFieldsOk() (*interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *UpdateDataViewRequestObjectDataView) SetFields(v interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *UpdateDataViewRequestObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### SetFieldsNil

`func (o *UpdateDataViewRequestObjectDataView) SetFieldsNil(b bool)`

 SetFieldsNil sets the value for Fields to be an explicit nil

### UnsetFields
`func (o *UpdateDataViewRequestObjectDataView) UnsetFields()`

UnsetFields ensures that no value is present for Fields, not even an explicit nil
### GetName

`func (o *UpdateDataViewRequestObjectDataView) GetName() interface{}`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateDataViewRequestObjectDataView) GetNameOk() (*interface{}, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateDataViewRequestObjectDataView) SetName(v interface{})`

SetName sets Name field to given value.

### HasName

`func (o *UpdateDataViewRequestObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *UpdateDataViewRequestObjectDataView) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *UpdateDataViewRequestObjectDataView) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMap() interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMapOk() (*interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) SetRuntimeFieldMap(v interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### SetRuntimeFieldMapNil

`func (o *UpdateDataViewRequestObjectDataView) SetRuntimeFieldMapNil(b bool)`

 SetRuntimeFieldMapNil sets the value for RuntimeFieldMap to be an explicit nil

### UnsetRuntimeFieldMap
`func (o *UpdateDataViewRequestObjectDataView) UnsetRuntimeFieldMap()`

UnsetRuntimeFieldMap ensures that no value is present for RuntimeFieldMap, not even an explicit nil
### GetSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) GetSourceFilters() interface{}`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *UpdateDataViewRequestObjectDataView) GetSourceFiltersOk() (*interface{}, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) SetSourceFilters(v interface{})`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### SetSourceFiltersNil

`func (o *UpdateDataViewRequestObjectDataView) SetSourceFiltersNil(b bool)`

 SetSourceFiltersNil sets the value for SourceFilters to be an explicit nil

### UnsetSourceFilters
`func (o *UpdateDataViewRequestObjectDataView) UnsetSourceFilters()`

UnsetSourceFilters ensures that no value is present for SourceFilters, not even an explicit nil
### GetTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldName() interface{}`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldNameOk() (*interface{}, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) SetTimeFieldName(v interface{})`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### SetTimeFieldNameNil

`func (o *UpdateDataViewRequestObjectDataView) SetTimeFieldNameNil(b bool)`

 SetTimeFieldNameNil sets the value for TimeFieldName to be an explicit nil

### UnsetTimeFieldName
`func (o *UpdateDataViewRequestObjectDataView) UnsetTimeFieldName()`

UnsetTimeFieldName ensures that no value is present for TimeFieldName, not even an explicit nil
### GetTitle

`func (o *UpdateDataViewRequestObjectDataView) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *UpdateDataViewRequestObjectDataView) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *UpdateDataViewRequestObjectDataView) SetTitle(v interface{})`

SetTitle sets Title field to given value.

### HasTitle

`func (o *UpdateDataViewRequestObjectDataView) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### SetTitleNil

`func (o *UpdateDataViewRequestObjectDataView) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *UpdateDataViewRequestObjectDataView) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil
### GetType

`func (o *UpdateDataViewRequestObjectDataView) GetType() interface{}`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UpdateDataViewRequestObjectDataView) GetTypeOk() (*interface{}, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UpdateDataViewRequestObjectDataView) SetType(v interface{})`

SetType sets Type field to given value.

### HasType

`func (o *UpdateDataViewRequestObjectDataView) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *UpdateDataViewRequestObjectDataView) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *UpdateDataViewRequestObjectDataView) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) GetTypeMeta() interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *UpdateDataViewRequestObjectDataView) GetTypeMetaOk() (*interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) SetTypeMeta(v interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.

### SetTypeMetaNil

`func (o *UpdateDataViewRequestObjectDataView) SetTypeMetaNil(b bool)`

 SetTypeMetaNil sets the value for TypeMeta to be an explicit nil

### UnsetTypeMeta
`func (o *UpdateDataViewRequestObjectDataView) UnsetTypeMeta()`

UnsetTypeMeta ensures that no value is present for TypeMeta, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


