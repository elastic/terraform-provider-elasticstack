# UpdateDataViewRequestObjectDataView

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowNoIndex** | Pointer to **bool** | Allows the data view saved object to exist before the data is available. | [optional] 
**FieldFormats** | Pointer to **map[string]interface{}** | A map of field formats by field name. | [optional] 
**Fields** | Pointer to **map[string]interface{}** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**RuntimeFieldMap** | Pointer to **map[string]interface{}** | A map of runtime field definitions by field name. | [optional] 
**SourceFilters** | Pointer to [**[]SourcefiltersInner**](SourcefiltersInner.md) | The array of field names you want to filter out in Discover. | [optional] 
**TimeFieldName** | Pointer to **string** | The timestamp field name, which you use for time-based data views. | [optional] 
**Title** | Pointer to **string** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | [optional] 
**Type** | Pointer to **string** | When set to &#x60;rollup&#x60;, identifies the rollup data views. | [optional] 
**TypeMeta** | Pointer to **map[string]interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 

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

`func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndex() bool`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *UpdateDataViewRequestObjectDataView) GetAllowNoIndexOk() (*bool, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *UpdateDataViewRequestObjectDataView) SetAllowNoIndex(v bool)`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *UpdateDataViewRequestObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### GetFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) GetFieldFormats() map[string]interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *UpdateDataViewRequestObjectDataView) GetFieldFormatsOk() (*map[string]interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) SetFieldFormats(v map[string]interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *UpdateDataViewRequestObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### GetFields

`func (o *UpdateDataViewRequestObjectDataView) GetFields() map[string]interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *UpdateDataViewRequestObjectDataView) GetFieldsOk() (*map[string]interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *UpdateDataViewRequestObjectDataView) SetFields(v map[string]interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *UpdateDataViewRequestObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetName

`func (o *UpdateDataViewRequestObjectDataView) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateDataViewRequestObjectDataView) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateDataViewRequestObjectDataView) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UpdateDataViewRequestObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### GetRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMap() map[string]interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *UpdateDataViewRequestObjectDataView) GetRuntimeFieldMapOk() (*map[string]interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) SetRuntimeFieldMap(v map[string]interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *UpdateDataViewRequestObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### GetSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) GetSourceFilters() []SourcefiltersInner`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *UpdateDataViewRequestObjectDataView) GetSourceFiltersOk() (*[]SourcefiltersInner, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) SetSourceFilters(v []SourcefiltersInner)`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *UpdateDataViewRequestObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### GetTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldName() string`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *UpdateDataViewRequestObjectDataView) GetTimeFieldNameOk() (*string, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) SetTimeFieldName(v string)`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *UpdateDataViewRequestObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### GetTitle

`func (o *UpdateDataViewRequestObjectDataView) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *UpdateDataViewRequestObjectDataView) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *UpdateDataViewRequestObjectDataView) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *UpdateDataViewRequestObjectDataView) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetType

`func (o *UpdateDataViewRequestObjectDataView) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *UpdateDataViewRequestObjectDataView) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *UpdateDataViewRequestObjectDataView) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *UpdateDataViewRequestObjectDataView) HasType() bool`

HasType returns a boolean if a field has been set.

### GetTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) GetTypeMeta() map[string]interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *UpdateDataViewRequestObjectDataView) GetTypeMetaOk() (*map[string]interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) SetTypeMeta(v map[string]interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *UpdateDataViewRequestObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


