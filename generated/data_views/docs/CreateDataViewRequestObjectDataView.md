# CreateDataViewRequestObjectDataView

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
**Title** | **string** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | 
**Type** | Pointer to **string** | When set to &#x60;rollup&#x60;, identifies the rollup data views. | [optional] 
**TypeMeta** | Pointer to **map[string]interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 
**Version** | Pointer to **string** |  | [optional] 

## Methods

### NewCreateDataViewRequestObjectDataView

`func NewCreateDataViewRequestObjectDataView(title string, ) *CreateDataViewRequestObjectDataView`

NewCreateDataViewRequestObjectDataView instantiates a new CreateDataViewRequestObjectDataView object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateDataViewRequestObjectDataViewWithDefaults

`func NewCreateDataViewRequestObjectDataViewWithDefaults() *CreateDataViewRequestObjectDataView`

NewCreateDataViewRequestObjectDataViewWithDefaults instantiates a new CreateDataViewRequestObjectDataView object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowNoIndex

`func (o *CreateDataViewRequestObjectDataView) GetAllowNoIndex() bool`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *CreateDataViewRequestObjectDataView) GetAllowNoIndexOk() (*bool, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *CreateDataViewRequestObjectDataView) SetAllowNoIndex(v bool)`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *CreateDataViewRequestObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### GetFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) GetFieldAttrs() map[string]interface{}`

GetFieldAttrs returns the FieldAttrs field if non-nil, zero value otherwise.

### GetFieldAttrsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldAttrsOk() (*map[string]interface{}, bool)`

GetFieldAttrsOk returns a tuple with the FieldAttrs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) SetFieldAttrs(v map[string]interface{})`

SetFieldAttrs sets FieldAttrs field to given value.

### HasFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) HasFieldAttrs() bool`

HasFieldAttrs returns a boolean if a field has been set.

### GetFieldFormats

`func (o *CreateDataViewRequestObjectDataView) GetFieldFormats() map[string]interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldFormatsOk() (*map[string]interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *CreateDataViewRequestObjectDataView) SetFieldFormats(v map[string]interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *CreateDataViewRequestObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### GetFields

`func (o *CreateDataViewRequestObjectDataView) GetFields() map[string]interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldsOk() (*map[string]interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *CreateDataViewRequestObjectDataView) SetFields(v map[string]interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *CreateDataViewRequestObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetId

`func (o *CreateDataViewRequestObjectDataView) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateDataViewRequestObjectDataView) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateDataViewRequestObjectDataView) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *CreateDataViewRequestObjectDataView) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *CreateDataViewRequestObjectDataView) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateDataViewRequestObjectDataView) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateDataViewRequestObjectDataView) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *CreateDataViewRequestObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNamespaces

`func (o *CreateDataViewRequestObjectDataView) GetNamespaces() []string`

GetNamespaces returns the Namespaces field if non-nil, zero value otherwise.

### GetNamespacesOk

`func (o *CreateDataViewRequestObjectDataView) GetNamespacesOk() (*[]string, bool)`

GetNamespacesOk returns a tuple with the Namespaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespaces

`func (o *CreateDataViewRequestObjectDataView) SetNamespaces(v []string)`

SetNamespaces sets Namespaces field to given value.

### HasNamespaces

`func (o *CreateDataViewRequestObjectDataView) HasNamespaces() bool`

HasNamespaces returns a boolean if a field has been set.

### GetRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) GetRuntimeFieldMap() map[string]interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *CreateDataViewRequestObjectDataView) GetRuntimeFieldMapOk() (*map[string]interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) SetRuntimeFieldMap(v map[string]interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### GetSourceFilters

`func (o *CreateDataViewRequestObjectDataView) GetSourceFilters() []SourcefiltersInner`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *CreateDataViewRequestObjectDataView) GetSourceFiltersOk() (*[]SourcefiltersInner, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *CreateDataViewRequestObjectDataView) SetSourceFilters(v []SourcefiltersInner)`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *CreateDataViewRequestObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### GetTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) GetTimeFieldName() string`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *CreateDataViewRequestObjectDataView) GetTimeFieldNameOk() (*string, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) SetTimeFieldName(v string)`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### GetTitle

`func (o *CreateDataViewRequestObjectDataView) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *CreateDataViewRequestObjectDataView) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *CreateDataViewRequestObjectDataView) SetTitle(v string)`

SetTitle sets Title field to given value.


### GetType

`func (o *CreateDataViewRequestObjectDataView) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *CreateDataViewRequestObjectDataView) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *CreateDataViewRequestObjectDataView) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *CreateDataViewRequestObjectDataView) HasType() bool`

HasType returns a boolean if a field has been set.

### GetTypeMeta

`func (o *CreateDataViewRequestObjectDataView) GetTypeMeta() map[string]interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *CreateDataViewRequestObjectDataView) GetTypeMetaOk() (*map[string]interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *CreateDataViewRequestObjectDataView) SetTypeMeta(v map[string]interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *CreateDataViewRequestObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.

### GetVersion

`func (o *CreateDataViewRequestObjectDataView) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *CreateDataViewRequestObjectDataView) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *CreateDataViewRequestObjectDataView) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *CreateDataViewRequestObjectDataView) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


