# DataViewResponseObjectDataView

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowNoIndex** | Pointer to **bool** | Allows the data view saved object to exist before the data is available. | [optional] 
**FieldAttrs** | Pointer to **map[string]interface{}** | A map of field attributes by field name. | [optional] 
**FieldFormats** | Pointer to **map[string]interface{}** | A map of field formats by field name. | [optional] 
**Fields** | Pointer to **map[string]interface{}** |  | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** | The data view name. | [optional] 
**Namespaces** | Pointer to **[]string** | An array of space identifiers for sharing the data view between multiple spaces. | [optional] 
**RuntimeFieldMap** | Pointer to **map[string]interface{}** | A map of runtime field definitions by field name. | [optional] 
**SourceFilters** | Pointer to [**[]SourcefiltersInner**](SourcefiltersInner.md) | The array of field names you want to filter out in Discover. | [optional] 
**TimeFieldName** | Pointer to **string** | The timestamp field name, which you use for time-based data views. | [optional] 
**Title** | Pointer to **string** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | [optional] 
**TypeMeta** | Pointer to **map[string]interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 
**Version** | Pointer to **string** |  | [optional] 

## Methods

### NewDataViewResponseObjectDataView

`func NewDataViewResponseObjectDataView() *DataViewResponseObjectDataView`

NewDataViewResponseObjectDataView instantiates a new DataViewResponseObjectDataView object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDataViewResponseObjectDataViewWithDefaults

`func NewDataViewResponseObjectDataViewWithDefaults() *DataViewResponseObjectDataView`

NewDataViewResponseObjectDataViewWithDefaults instantiates a new DataViewResponseObjectDataView object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowNoIndex

`func (o *DataViewResponseObjectDataView) GetAllowNoIndex() bool`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *DataViewResponseObjectDataView) GetAllowNoIndexOk() (*bool, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *DataViewResponseObjectDataView) SetAllowNoIndex(v bool)`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *DataViewResponseObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### GetFieldAttrs

`func (o *DataViewResponseObjectDataView) GetFieldAttrs() map[string]interface{}`

GetFieldAttrs returns the FieldAttrs field if non-nil, zero value otherwise.

### GetFieldAttrsOk

`func (o *DataViewResponseObjectDataView) GetFieldAttrsOk() (*map[string]interface{}, bool)`

GetFieldAttrsOk returns a tuple with the FieldAttrs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldAttrs

`func (o *DataViewResponseObjectDataView) SetFieldAttrs(v map[string]interface{})`

SetFieldAttrs sets FieldAttrs field to given value.

### HasFieldAttrs

`func (o *DataViewResponseObjectDataView) HasFieldAttrs() bool`

HasFieldAttrs returns a boolean if a field has been set.

### GetFieldFormats

`func (o *DataViewResponseObjectDataView) GetFieldFormats() map[string]interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *DataViewResponseObjectDataView) GetFieldFormatsOk() (*map[string]interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *DataViewResponseObjectDataView) SetFieldFormats(v map[string]interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *DataViewResponseObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### GetFields

`func (o *DataViewResponseObjectDataView) GetFields() map[string]interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *DataViewResponseObjectDataView) GetFieldsOk() (*map[string]interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *DataViewResponseObjectDataView) SetFields(v map[string]interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *DataViewResponseObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetId

`func (o *DataViewResponseObjectDataView) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DataViewResponseObjectDataView) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DataViewResponseObjectDataView) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *DataViewResponseObjectDataView) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *DataViewResponseObjectDataView) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DataViewResponseObjectDataView) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DataViewResponseObjectDataView) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DataViewResponseObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNamespaces

`func (o *DataViewResponseObjectDataView) GetNamespaces() []string`

GetNamespaces returns the Namespaces field if non-nil, zero value otherwise.

### GetNamespacesOk

`func (o *DataViewResponseObjectDataView) GetNamespacesOk() (*[]string, bool)`

GetNamespacesOk returns a tuple with the Namespaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespaces

`func (o *DataViewResponseObjectDataView) SetNamespaces(v []string)`

SetNamespaces sets Namespaces field to given value.

### HasNamespaces

`func (o *DataViewResponseObjectDataView) HasNamespaces() bool`

HasNamespaces returns a boolean if a field has been set.

### GetRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) GetRuntimeFieldMap() map[string]interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *DataViewResponseObjectDataView) GetRuntimeFieldMapOk() (*map[string]interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) SetRuntimeFieldMap(v map[string]interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### GetSourceFilters

`func (o *DataViewResponseObjectDataView) GetSourceFilters() []SourcefiltersInner`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *DataViewResponseObjectDataView) GetSourceFiltersOk() (*[]SourcefiltersInner, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *DataViewResponseObjectDataView) SetSourceFilters(v []SourcefiltersInner)`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *DataViewResponseObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### GetTimeFieldName

`func (o *DataViewResponseObjectDataView) GetTimeFieldName() string`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *DataViewResponseObjectDataView) GetTimeFieldNameOk() (*string, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *DataViewResponseObjectDataView) SetTimeFieldName(v string)`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *DataViewResponseObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### GetTitle

`func (o *DataViewResponseObjectDataView) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *DataViewResponseObjectDataView) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *DataViewResponseObjectDataView) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *DataViewResponseObjectDataView) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetTypeMeta

`func (o *DataViewResponseObjectDataView) GetTypeMeta() map[string]interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *DataViewResponseObjectDataView) GetTypeMetaOk() (*map[string]interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *DataViewResponseObjectDataView) SetTypeMeta(v map[string]interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *DataViewResponseObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.

### GetVersion

`func (o *DataViewResponseObjectDataView) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *DataViewResponseObjectDataView) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *DataViewResponseObjectDataView) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *DataViewResponseObjectDataView) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


