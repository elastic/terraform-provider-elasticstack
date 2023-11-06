# DataViewResponseObjectDataView

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowNoIndex** | Pointer to **interface{}** | Allows the data view saved object to exist before the data is available. | [optional] 
**FieldAttrs** | Pointer to **interface{}** | A map of field attributes by field name. | [optional] 
**FieldFormats** | Pointer to **interface{}** | A map of field formats by field name. | [optional] 
**Fields** | Pointer to **interface{}** |  | [optional] 
**Id** | Pointer to **interface{}** |  | [optional] 
**Name** | Pointer to **interface{}** | The data view name. | [optional] 
**Namespaces** | Pointer to **interface{}** | An array of space identifiers for sharing the data view between multiple spaces. | [optional] 
**RuntimeFieldMap** | Pointer to **interface{}** | A map of runtime field definitions by field name. | [optional] 
**SourceFilters** | Pointer to **interface{}** | The array of field names you want to filter out in Discover. | [optional] 
**TimeFieldName** | Pointer to **interface{}** | The timestamp field name, which you use for time-based data views. | [optional] 
**Title** | Pointer to **interface{}** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | [optional] 
**TypeMeta** | Pointer to **interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 
**Version** | Pointer to **interface{}** |  | [optional] 

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

`func (o *DataViewResponseObjectDataView) GetAllowNoIndex() interface{}`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *DataViewResponseObjectDataView) GetAllowNoIndexOk() (*interface{}, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *DataViewResponseObjectDataView) SetAllowNoIndex(v interface{})`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *DataViewResponseObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### SetAllowNoIndexNil

`func (o *DataViewResponseObjectDataView) SetAllowNoIndexNil(b bool)`

 SetAllowNoIndexNil sets the value for AllowNoIndex to be an explicit nil

### UnsetAllowNoIndex
`func (o *DataViewResponseObjectDataView) UnsetAllowNoIndex()`

UnsetAllowNoIndex ensures that no value is present for AllowNoIndex, not even an explicit nil
### GetFieldAttrs

`func (o *DataViewResponseObjectDataView) GetFieldAttrs() interface{}`

GetFieldAttrs returns the FieldAttrs field if non-nil, zero value otherwise.

### GetFieldAttrsOk

`func (o *DataViewResponseObjectDataView) GetFieldAttrsOk() (*interface{}, bool)`

GetFieldAttrsOk returns a tuple with the FieldAttrs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldAttrs

`func (o *DataViewResponseObjectDataView) SetFieldAttrs(v interface{})`

SetFieldAttrs sets FieldAttrs field to given value.

### HasFieldAttrs

`func (o *DataViewResponseObjectDataView) HasFieldAttrs() bool`

HasFieldAttrs returns a boolean if a field has been set.

### SetFieldAttrsNil

`func (o *DataViewResponseObjectDataView) SetFieldAttrsNil(b bool)`

 SetFieldAttrsNil sets the value for FieldAttrs to be an explicit nil

### UnsetFieldAttrs
`func (o *DataViewResponseObjectDataView) UnsetFieldAttrs()`

UnsetFieldAttrs ensures that no value is present for FieldAttrs, not even an explicit nil
### GetFieldFormats

`func (o *DataViewResponseObjectDataView) GetFieldFormats() interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *DataViewResponseObjectDataView) GetFieldFormatsOk() (*interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *DataViewResponseObjectDataView) SetFieldFormats(v interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *DataViewResponseObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### SetFieldFormatsNil

`func (o *DataViewResponseObjectDataView) SetFieldFormatsNil(b bool)`

 SetFieldFormatsNil sets the value for FieldFormats to be an explicit nil

### UnsetFieldFormats
`func (o *DataViewResponseObjectDataView) UnsetFieldFormats()`

UnsetFieldFormats ensures that no value is present for FieldFormats, not even an explicit nil
### GetFields

`func (o *DataViewResponseObjectDataView) GetFields() interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *DataViewResponseObjectDataView) GetFieldsOk() (*interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *DataViewResponseObjectDataView) SetFields(v interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *DataViewResponseObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### SetFieldsNil

`func (o *DataViewResponseObjectDataView) SetFieldsNil(b bool)`

 SetFieldsNil sets the value for Fields to be an explicit nil

### UnsetFields
`func (o *DataViewResponseObjectDataView) UnsetFields()`

UnsetFields ensures that no value is present for Fields, not even an explicit nil
### GetId

`func (o *DataViewResponseObjectDataView) GetId() interface{}`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DataViewResponseObjectDataView) GetIdOk() (*interface{}, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DataViewResponseObjectDataView) SetId(v interface{})`

SetId sets Id field to given value.

### HasId

`func (o *DataViewResponseObjectDataView) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *DataViewResponseObjectDataView) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *DataViewResponseObjectDataView) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetName

`func (o *DataViewResponseObjectDataView) GetName() interface{}`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DataViewResponseObjectDataView) GetNameOk() (*interface{}, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DataViewResponseObjectDataView) SetName(v interface{})`

SetName sets Name field to given value.

### HasName

`func (o *DataViewResponseObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *DataViewResponseObjectDataView) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *DataViewResponseObjectDataView) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetNamespaces

`func (o *DataViewResponseObjectDataView) GetNamespaces() interface{}`

GetNamespaces returns the Namespaces field if non-nil, zero value otherwise.

### GetNamespacesOk

`func (o *DataViewResponseObjectDataView) GetNamespacesOk() (*interface{}, bool)`

GetNamespacesOk returns a tuple with the Namespaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespaces

`func (o *DataViewResponseObjectDataView) SetNamespaces(v interface{})`

SetNamespaces sets Namespaces field to given value.

### HasNamespaces

`func (o *DataViewResponseObjectDataView) HasNamespaces() bool`

HasNamespaces returns a boolean if a field has been set.

### SetNamespacesNil

`func (o *DataViewResponseObjectDataView) SetNamespacesNil(b bool)`

 SetNamespacesNil sets the value for Namespaces to be an explicit nil

### UnsetNamespaces
`func (o *DataViewResponseObjectDataView) UnsetNamespaces()`

UnsetNamespaces ensures that no value is present for Namespaces, not even an explicit nil
### GetRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) GetRuntimeFieldMap() interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *DataViewResponseObjectDataView) GetRuntimeFieldMapOk() (*interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) SetRuntimeFieldMap(v interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *DataViewResponseObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### SetRuntimeFieldMapNil

`func (o *DataViewResponseObjectDataView) SetRuntimeFieldMapNil(b bool)`

 SetRuntimeFieldMapNil sets the value for RuntimeFieldMap to be an explicit nil

### UnsetRuntimeFieldMap
`func (o *DataViewResponseObjectDataView) UnsetRuntimeFieldMap()`

UnsetRuntimeFieldMap ensures that no value is present for RuntimeFieldMap, not even an explicit nil
### GetSourceFilters

`func (o *DataViewResponseObjectDataView) GetSourceFilters() interface{}`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *DataViewResponseObjectDataView) GetSourceFiltersOk() (*interface{}, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *DataViewResponseObjectDataView) SetSourceFilters(v interface{})`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *DataViewResponseObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### SetSourceFiltersNil

`func (o *DataViewResponseObjectDataView) SetSourceFiltersNil(b bool)`

 SetSourceFiltersNil sets the value for SourceFilters to be an explicit nil

### UnsetSourceFilters
`func (o *DataViewResponseObjectDataView) UnsetSourceFilters()`

UnsetSourceFilters ensures that no value is present for SourceFilters, not even an explicit nil
### GetTimeFieldName

`func (o *DataViewResponseObjectDataView) GetTimeFieldName() interface{}`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *DataViewResponseObjectDataView) GetTimeFieldNameOk() (*interface{}, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *DataViewResponseObjectDataView) SetTimeFieldName(v interface{})`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *DataViewResponseObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### SetTimeFieldNameNil

`func (o *DataViewResponseObjectDataView) SetTimeFieldNameNil(b bool)`

 SetTimeFieldNameNil sets the value for TimeFieldName to be an explicit nil

### UnsetTimeFieldName
`func (o *DataViewResponseObjectDataView) UnsetTimeFieldName()`

UnsetTimeFieldName ensures that no value is present for TimeFieldName, not even an explicit nil
### GetTitle

`func (o *DataViewResponseObjectDataView) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *DataViewResponseObjectDataView) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *DataViewResponseObjectDataView) SetTitle(v interface{})`

SetTitle sets Title field to given value.

### HasTitle

`func (o *DataViewResponseObjectDataView) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### SetTitleNil

`func (o *DataViewResponseObjectDataView) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *DataViewResponseObjectDataView) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil
### GetTypeMeta

`func (o *DataViewResponseObjectDataView) GetTypeMeta() interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *DataViewResponseObjectDataView) GetTypeMetaOk() (*interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *DataViewResponseObjectDataView) SetTypeMeta(v interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *DataViewResponseObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.

### SetTypeMetaNil

`func (o *DataViewResponseObjectDataView) SetTypeMetaNil(b bool)`

 SetTypeMetaNil sets the value for TypeMeta to be an explicit nil

### UnsetTypeMeta
`func (o *DataViewResponseObjectDataView) UnsetTypeMeta()`

UnsetTypeMeta ensures that no value is present for TypeMeta, not even an explicit nil
### GetVersion

`func (o *DataViewResponseObjectDataView) GetVersion() interface{}`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *DataViewResponseObjectDataView) GetVersionOk() (*interface{}, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *DataViewResponseObjectDataView) SetVersion(v interface{})`

SetVersion sets Version field to given value.

### HasVersion

`func (o *DataViewResponseObjectDataView) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### SetVersionNil

`func (o *DataViewResponseObjectDataView) SetVersionNil(b bool)`

 SetVersionNil sets the value for Version to be an explicit nil

### UnsetVersion
`func (o *DataViewResponseObjectDataView) UnsetVersion()`

UnsetVersion ensures that no value is present for Version, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


