# CreateDataViewRequestObjectDataView

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
**Title** | **interface{}** | Comma-separated list of data streams, indices, and aliases that you want to search. Supports wildcards (&#x60;*&#x60;). | 
**Type** | Pointer to **interface{}** | When set to &#x60;rollup&#x60;, identifies the rollup data views. | [optional] 
**TypeMeta** | Pointer to **interface{}** | When you use rollup indices, contains the field list for the rollup data view API endpoints. | [optional] 
**Version** | Pointer to **interface{}** |  | [optional] 

## Methods

### NewCreateDataViewRequestObjectDataView

`func NewCreateDataViewRequestObjectDataView(title interface{}, ) *CreateDataViewRequestObjectDataView`

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

`func (o *CreateDataViewRequestObjectDataView) GetAllowNoIndex() interface{}`

GetAllowNoIndex returns the AllowNoIndex field if non-nil, zero value otherwise.

### GetAllowNoIndexOk

`func (o *CreateDataViewRequestObjectDataView) GetAllowNoIndexOk() (*interface{}, bool)`

GetAllowNoIndexOk returns a tuple with the AllowNoIndex field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowNoIndex

`func (o *CreateDataViewRequestObjectDataView) SetAllowNoIndex(v interface{})`

SetAllowNoIndex sets AllowNoIndex field to given value.

### HasAllowNoIndex

`func (o *CreateDataViewRequestObjectDataView) HasAllowNoIndex() bool`

HasAllowNoIndex returns a boolean if a field has been set.

### SetAllowNoIndexNil

`func (o *CreateDataViewRequestObjectDataView) SetAllowNoIndexNil(b bool)`

 SetAllowNoIndexNil sets the value for AllowNoIndex to be an explicit nil

### UnsetAllowNoIndex
`func (o *CreateDataViewRequestObjectDataView) UnsetAllowNoIndex()`

UnsetAllowNoIndex ensures that no value is present for AllowNoIndex, not even an explicit nil
### GetFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) GetFieldAttrs() interface{}`

GetFieldAttrs returns the FieldAttrs field if non-nil, zero value otherwise.

### GetFieldAttrsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldAttrsOk() (*interface{}, bool)`

GetFieldAttrsOk returns a tuple with the FieldAttrs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) SetFieldAttrs(v interface{})`

SetFieldAttrs sets FieldAttrs field to given value.

### HasFieldAttrs

`func (o *CreateDataViewRequestObjectDataView) HasFieldAttrs() bool`

HasFieldAttrs returns a boolean if a field has been set.

### SetFieldAttrsNil

`func (o *CreateDataViewRequestObjectDataView) SetFieldAttrsNil(b bool)`

 SetFieldAttrsNil sets the value for FieldAttrs to be an explicit nil

### UnsetFieldAttrs
`func (o *CreateDataViewRequestObjectDataView) UnsetFieldAttrs()`

UnsetFieldAttrs ensures that no value is present for FieldAttrs, not even an explicit nil
### GetFieldFormats

`func (o *CreateDataViewRequestObjectDataView) GetFieldFormats() interface{}`

GetFieldFormats returns the FieldFormats field if non-nil, zero value otherwise.

### GetFieldFormatsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldFormatsOk() (*interface{}, bool)`

GetFieldFormatsOk returns a tuple with the FieldFormats field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldFormats

`func (o *CreateDataViewRequestObjectDataView) SetFieldFormats(v interface{})`

SetFieldFormats sets FieldFormats field to given value.

### HasFieldFormats

`func (o *CreateDataViewRequestObjectDataView) HasFieldFormats() bool`

HasFieldFormats returns a boolean if a field has been set.

### SetFieldFormatsNil

`func (o *CreateDataViewRequestObjectDataView) SetFieldFormatsNil(b bool)`

 SetFieldFormatsNil sets the value for FieldFormats to be an explicit nil

### UnsetFieldFormats
`func (o *CreateDataViewRequestObjectDataView) UnsetFieldFormats()`

UnsetFieldFormats ensures that no value is present for FieldFormats, not even an explicit nil
### GetFields

`func (o *CreateDataViewRequestObjectDataView) GetFields() interface{}`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *CreateDataViewRequestObjectDataView) GetFieldsOk() (*interface{}, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *CreateDataViewRequestObjectDataView) SetFields(v interface{})`

SetFields sets Fields field to given value.

### HasFields

`func (o *CreateDataViewRequestObjectDataView) HasFields() bool`

HasFields returns a boolean if a field has been set.

### SetFieldsNil

`func (o *CreateDataViewRequestObjectDataView) SetFieldsNil(b bool)`

 SetFieldsNil sets the value for Fields to be an explicit nil

### UnsetFields
`func (o *CreateDataViewRequestObjectDataView) UnsetFields()`

UnsetFields ensures that no value is present for Fields, not even an explicit nil
### GetId

`func (o *CreateDataViewRequestObjectDataView) GetId() interface{}`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *CreateDataViewRequestObjectDataView) GetIdOk() (*interface{}, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *CreateDataViewRequestObjectDataView) SetId(v interface{})`

SetId sets Id field to given value.

### HasId

`func (o *CreateDataViewRequestObjectDataView) HasId() bool`

HasId returns a boolean if a field has been set.

### SetIdNil

`func (o *CreateDataViewRequestObjectDataView) SetIdNil(b bool)`

 SetIdNil sets the value for Id to be an explicit nil

### UnsetId
`func (o *CreateDataViewRequestObjectDataView) UnsetId()`

UnsetId ensures that no value is present for Id, not even an explicit nil
### GetName

`func (o *CreateDataViewRequestObjectDataView) GetName() interface{}`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateDataViewRequestObjectDataView) GetNameOk() (*interface{}, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateDataViewRequestObjectDataView) SetName(v interface{})`

SetName sets Name field to given value.

### HasName

`func (o *CreateDataViewRequestObjectDataView) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *CreateDataViewRequestObjectDataView) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *CreateDataViewRequestObjectDataView) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetNamespaces

`func (o *CreateDataViewRequestObjectDataView) GetNamespaces() interface{}`

GetNamespaces returns the Namespaces field if non-nil, zero value otherwise.

### GetNamespacesOk

`func (o *CreateDataViewRequestObjectDataView) GetNamespacesOk() (*interface{}, bool)`

GetNamespacesOk returns a tuple with the Namespaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespaces

`func (o *CreateDataViewRequestObjectDataView) SetNamespaces(v interface{})`

SetNamespaces sets Namespaces field to given value.

### HasNamespaces

`func (o *CreateDataViewRequestObjectDataView) HasNamespaces() bool`

HasNamespaces returns a boolean if a field has been set.

### SetNamespacesNil

`func (o *CreateDataViewRequestObjectDataView) SetNamespacesNil(b bool)`

 SetNamespacesNil sets the value for Namespaces to be an explicit nil

### UnsetNamespaces
`func (o *CreateDataViewRequestObjectDataView) UnsetNamespaces()`

UnsetNamespaces ensures that no value is present for Namespaces, not even an explicit nil
### GetRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) GetRuntimeFieldMap() interface{}`

GetRuntimeFieldMap returns the RuntimeFieldMap field if non-nil, zero value otherwise.

### GetRuntimeFieldMapOk

`func (o *CreateDataViewRequestObjectDataView) GetRuntimeFieldMapOk() (*interface{}, bool)`

GetRuntimeFieldMapOk returns a tuple with the RuntimeFieldMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) SetRuntimeFieldMap(v interface{})`

SetRuntimeFieldMap sets RuntimeFieldMap field to given value.

### HasRuntimeFieldMap

`func (o *CreateDataViewRequestObjectDataView) HasRuntimeFieldMap() bool`

HasRuntimeFieldMap returns a boolean if a field has been set.

### SetRuntimeFieldMapNil

`func (o *CreateDataViewRequestObjectDataView) SetRuntimeFieldMapNil(b bool)`

 SetRuntimeFieldMapNil sets the value for RuntimeFieldMap to be an explicit nil

### UnsetRuntimeFieldMap
`func (o *CreateDataViewRequestObjectDataView) UnsetRuntimeFieldMap()`

UnsetRuntimeFieldMap ensures that no value is present for RuntimeFieldMap, not even an explicit nil
### GetSourceFilters

`func (o *CreateDataViewRequestObjectDataView) GetSourceFilters() interface{}`

GetSourceFilters returns the SourceFilters field if non-nil, zero value otherwise.

### GetSourceFiltersOk

`func (o *CreateDataViewRequestObjectDataView) GetSourceFiltersOk() (*interface{}, bool)`

GetSourceFiltersOk returns a tuple with the SourceFilters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceFilters

`func (o *CreateDataViewRequestObjectDataView) SetSourceFilters(v interface{})`

SetSourceFilters sets SourceFilters field to given value.

### HasSourceFilters

`func (o *CreateDataViewRequestObjectDataView) HasSourceFilters() bool`

HasSourceFilters returns a boolean if a field has been set.

### SetSourceFiltersNil

`func (o *CreateDataViewRequestObjectDataView) SetSourceFiltersNil(b bool)`

 SetSourceFiltersNil sets the value for SourceFilters to be an explicit nil

### UnsetSourceFilters
`func (o *CreateDataViewRequestObjectDataView) UnsetSourceFilters()`

UnsetSourceFilters ensures that no value is present for SourceFilters, not even an explicit nil
### GetTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) GetTimeFieldName() interface{}`

GetTimeFieldName returns the TimeFieldName field if non-nil, zero value otherwise.

### GetTimeFieldNameOk

`func (o *CreateDataViewRequestObjectDataView) GetTimeFieldNameOk() (*interface{}, bool)`

GetTimeFieldNameOk returns a tuple with the TimeFieldName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) SetTimeFieldName(v interface{})`

SetTimeFieldName sets TimeFieldName field to given value.

### HasTimeFieldName

`func (o *CreateDataViewRequestObjectDataView) HasTimeFieldName() bool`

HasTimeFieldName returns a boolean if a field has been set.

### SetTimeFieldNameNil

`func (o *CreateDataViewRequestObjectDataView) SetTimeFieldNameNil(b bool)`

 SetTimeFieldNameNil sets the value for TimeFieldName to be an explicit nil

### UnsetTimeFieldName
`func (o *CreateDataViewRequestObjectDataView) UnsetTimeFieldName()`

UnsetTimeFieldName ensures that no value is present for TimeFieldName, not even an explicit nil
### GetTitle

`func (o *CreateDataViewRequestObjectDataView) GetTitle() interface{}`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *CreateDataViewRequestObjectDataView) GetTitleOk() (*interface{}, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *CreateDataViewRequestObjectDataView) SetTitle(v interface{})`

SetTitle sets Title field to given value.


### SetTitleNil

`func (o *CreateDataViewRequestObjectDataView) SetTitleNil(b bool)`

 SetTitleNil sets the value for Title to be an explicit nil

### UnsetTitle
`func (o *CreateDataViewRequestObjectDataView) UnsetTitle()`

UnsetTitle ensures that no value is present for Title, not even an explicit nil
### GetType

`func (o *CreateDataViewRequestObjectDataView) GetType() interface{}`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *CreateDataViewRequestObjectDataView) GetTypeOk() (*interface{}, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *CreateDataViewRequestObjectDataView) SetType(v interface{})`

SetType sets Type field to given value.

### HasType

`func (o *CreateDataViewRequestObjectDataView) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *CreateDataViewRequestObjectDataView) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *CreateDataViewRequestObjectDataView) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetTypeMeta

`func (o *CreateDataViewRequestObjectDataView) GetTypeMeta() interface{}`

GetTypeMeta returns the TypeMeta field if non-nil, zero value otherwise.

### GetTypeMetaOk

`func (o *CreateDataViewRequestObjectDataView) GetTypeMetaOk() (*interface{}, bool)`

GetTypeMetaOk returns a tuple with the TypeMeta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTypeMeta

`func (o *CreateDataViewRequestObjectDataView) SetTypeMeta(v interface{})`

SetTypeMeta sets TypeMeta field to given value.

### HasTypeMeta

`func (o *CreateDataViewRequestObjectDataView) HasTypeMeta() bool`

HasTypeMeta returns a boolean if a field has been set.

### SetTypeMetaNil

`func (o *CreateDataViewRequestObjectDataView) SetTypeMetaNil(b bool)`

 SetTypeMetaNil sets the value for TypeMeta to be an explicit nil

### UnsetTypeMeta
`func (o *CreateDataViewRequestObjectDataView) UnsetTypeMeta()`

UnsetTypeMeta ensures that no value is present for TypeMeta, not even an explicit nil
### GetVersion

`func (o *CreateDataViewRequestObjectDataView) GetVersion() interface{}`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *CreateDataViewRequestObjectDataView) GetVersionOk() (*interface{}, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *CreateDataViewRequestObjectDataView) SetVersion(v interface{})`

SetVersion sets Version field to given value.

### HasVersion

`func (o *CreateDataViewRequestObjectDataView) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### SetVersionNil

`func (o *CreateDataViewRequestObjectDataView) SetVersionNil(b bool)`

 SetVersionNil sets the value for Version to be an explicit nil

### UnsetVersion
`func (o *CreateDataViewRequestObjectDataView) UnsetVersion()`

UnsetVersion ensures that no value is present for Version, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


