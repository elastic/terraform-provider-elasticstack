package kbapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	basePathKibanaSavedObject = "/api/saved_objects" // Base URL to access on Kibana save objects
)

// OptionalFindParameters contain optional parameters to find objects
type OptionalFindParameters struct {
	ObjectsPerPage        int
	Page                  int
	Search                string
	DefaultSearchOperator string
	SearchFields          []string
	Fields                []string
	SortField             string
	HasReference          string
}

// KibanaSavedObjectGet permit to get saved object from Kibana
type KibanaSavedObjectGet func(objectType string, id string, kibanaSpace string) (map[string]interface{}, error)

// KibanaSavedObjectFind permit to find saved objects from Kibana
type KibanaSavedObjectFind func(objectType string, kibanaSpace string, optionalParameters *OptionalFindParameters) (map[string]interface{}, error)

// KibanaSavedObjectCreate permit to create saved object in Kibana
type KibanaSavedObjectCreate func(data map[string]interface{}, objectType string, id string, overwrite bool, kibanaSpace string) (map[string]interface{}, error)

// KibanaSavedObjectUpdate permit to update saved object in Kibana
type KibanaSavedObjectUpdate func(data map[string]interface{}, objectType string, id string, kibanaSpace string) (map[string]interface{}, error)

// KibanaSavedObjectDelete permit to delete saved object in Kibana
type KibanaSavedObjectDelete func(objectType string, id string, kibanaSpace string) error

// KibanaSavedObjectExport permit to export saved objects from Kibana
type KibanaSavedObjectExport func(objectTypes []string, objects []map[string]string, deepReference bool, kibanaSpace string) ([]byte, error)

// KibanaSavedObjectImport permit to import saved objects in Kibana
type KibanaSavedObjectImport func(data []byte, overwrite bool, kibanaSpace string) (map[string]interface{}, error)

// String permit to return OptionalFindParameters object as JSON string
func (o *OptionalFindParameters) String() string {
	json, _ := json.Marshal(o)
	return string(json)
}

// newKibanaSavedObjectGetFunc permit to get saved obejct by it id and type
func newKibanaSavedObjectGetFunc(c *resty.Client) KibanaSavedObjectGet {
	return func(objectType string, id string, kibanaSpace string) (map[string]interface{}, error) {

		if objectType == "" {
			return nil, NewAPIError(600, "You must provide the object type")
		}
		if id == "" {
			return nil, NewAPIError(600, "You must provide the object ID")
		}
		log.Debug("ObjectType: ", objectType)
		log.Debug("ID: ", id)
		log.Debug("KibanaSpace: ", kibanaSpace)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/%s/%s", basePathKibanaSavedObject, objectType, id)
		} else {
			path = fmt.Sprintf("/s/%s%s/%s/%s", kibanaSpace, basePathKibanaSavedObject, objectType, id)
		}
		log.Debugf("URL to get object: %s", path)

		resp, err := c.R().Get(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			if resp.StatusCode() == 404 {
				return nil, nil
			}
			return nil, NewAPIError(resp.StatusCode(), resp.Status())

		}
		var data map[string]interface{}
		err = json.Unmarshal(resp.Body(), &data)
		if err != nil {
			return nil, err
		}
		log.Debug("Data: ", data)

		return data, nil
	}

}

// newKibanaSavedObjectFindFunc permit to search objects
func newKibanaSavedObjectFindFunc(c *resty.Client) KibanaSavedObjectFind {
	return func(objectType string, kibanaSpace string, optionalParameters *OptionalFindParameters) (map[string]interface{}, error) {

		if objectType == "" {
			return nil, NewAPIError(600, "You must provide the object type")
		}
		log.Debug("ObjectType: ", objectType)
		log.Debug("KibanaSpace : ", kibanaSpace)

		queryParams := map[string]string{
			"type": objectType,
		}
		if optionalParameters != nil {
			log.Debug("Objects Per Page: ", optionalParameters.ObjectsPerPage)
			log.Debug("Page: ", optionalParameters.Page)
			log.Debug("Search: ", optionalParameters.Search)
			log.Debug("DefaultSearchOperator: ", optionalParameters.DefaultSearchOperator)
			log.Debug("SearchFields: ", optionalParameters.SearchFields)
			log.Debug("Fields: ", optionalParameters.Fields)
			log.Debug("SortField: ", optionalParameters.SortField)
			log.Debug("HasReference: ", optionalParameters.HasReference)
			if optionalParameters.ObjectsPerPage != 0 {
				queryParams["per_page"] = strconv.Itoa(optionalParameters.ObjectsPerPage)
			}
			if optionalParameters.Page != 0 {
				queryParams["page"] = strconv.Itoa(optionalParameters.Page)
			}
			if optionalParameters.Search != "" {
				queryParams["search"] = optionalParameters.Search
			}
			if optionalParameters.DefaultSearchOperator != "" {
				queryParams["default_search_operator"] = optionalParameters.DefaultSearchOperator
			}
			if optionalParameters.SearchFields != nil {
				queryParams["search_fields"] = strings.Join(optionalParameters.SearchFields, ",")
			}
			if optionalParameters.Fields != nil {
				queryParams["fields"] = strings.Join(optionalParameters.Fields, ",")
			}
			if optionalParameters.SortField != "" {
				queryParams["sort_field"] = optionalParameters.SortField
			}
			if optionalParameters.HasReference != "" {
				queryParams["has_reference"] = optionalParameters.HasReference
			}
		}

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/_find", basePathKibanaSavedObject)
		} else {
			path = fmt.Sprintf("/s/%s%s/_find", kibanaSpace, basePathKibanaSavedObject)
		}
		log.Debugf("URL to find object: %s", path)

		resp, err := c.R().SetQueryParams(queryParams).Get(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			if resp.StatusCode() == 404 {
				return nil, nil
			}
			return nil, NewAPIError(resp.StatusCode(), resp.Status())

		}
		var data map[string]interface{}
		err = json.Unmarshal(resp.Body(), &data)
		if err != nil {
			return nil, err
		}
		log.Debug("Data: ", data)

		return data, nil
	}

}

// newKibanaSavedObjectCreateFunc permit to create new object on Kibana
func newKibanaSavedObjectCreateFunc(c *resty.Client) KibanaSavedObjectCreate {
	return func(data map[string]interface{}, objectType string, id string, overwrite bool, kibanaSpace string) (map[string]interface{}, error) {

		if data == nil {
			return nil, NewAPIError(600, "You must provide one or more dashboard to import")
		}
		if objectType == "" {
			return nil, NewAPIError(600, "You must provide the object type")
		}
		log.Debug("data: ", data)
		log.Debug("ObjectType: ", objectType)
		log.Debug("ID: ", id)
		log.Debug("Overwrite: ", overwrite)
		log.Debug("KibanaSpace: ", kibanaSpace)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/%s/%s", basePathKibanaSavedObject, objectType, id)
		} else {
			path = fmt.Sprintf("/s/%s%s/%s/%s", kibanaSpace, basePathKibanaSavedObject, objectType, id)
		}
		log.Debugf("URL to create object: %s", path)

		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		resp, err := c.R().SetQueryString(fmt.Sprintf("overwrite=%t", overwrite)).SetBody(jsonData).Post(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		var dataResponse map[string]interface{}
		err = json.Unmarshal(resp.Body(), &dataResponse)
		if err != nil {
			return nil, err
		}
		log.Debug("Data response: ", dataResponse)

		return dataResponse, nil
	}
}

// newKibanaSavedObjectUpdateFunc permit to update object on Kibana
func newKibanaSavedObjectUpdateFunc(c *resty.Client) KibanaSavedObjectUpdate {
	return func(data map[string]interface{}, objectType string, id string, kibanaSpace string) (map[string]interface{}, error) {

		if data == nil {
			return nil, NewAPIError(600, "You must provide one or more dashboard to import")
		}
		if objectType == "" {
			return nil, NewAPIError(600, "You must provide the object type")
		}
		if id == "" {
			return nil, NewAPIError(600, "You must provide the ID")
		}
		log.Debug("data: ", data)
		log.Debug("ObjectType: ", objectType)
		log.Debug("ID: ", id)
		log.Debug("kibanaSpace: ", kibanaSpace)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/%s/%s", basePathKibanaSavedObject, objectType, id)
		} else {
			path = fmt.Sprintf("/s/%s%s/%s/%s", kibanaSpace, basePathKibanaSavedObject, objectType, id)
		}
		log.Debugf("URL to update object: %s", path)

		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		resp, err := c.R().SetBody(jsonData).Put(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		var dataResponse map[string]interface{}
		err = json.Unmarshal(resp.Body(), &dataResponse)
		if err != nil {
			return nil, err
		}
		log.Debug("Data response: ", dataResponse)

		return dataResponse, nil
	}
}

// newKibanaSavedObjectDeleteFunc permit to delete object on Kibana
func newKibanaSavedObjectDeleteFunc(c *resty.Client) KibanaSavedObjectDelete {
	return func(objectType string, id string, kibanaSpace string) error {

		if objectType == "" {
			return NewAPIError(600, "You must provide the object type")
		}
		if id == "" {
			return NewAPIError(600, "You must provide the id")
		}
		log.Debug("objectType: ", objectType)
		log.Debug("ID: ", id)
		log.Debug("KibanaSpace: ", kibanaSpace)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/%s/%s", basePathKibanaSavedObject, objectType, id)
		} else {
			path = fmt.Sprintf("/s/%s%s/%s/%s", kibanaSpace, basePathKibanaSavedObject, objectType, id)
		}
		log.Debugf("URL to delete object: %s", path)

		resp, err := c.R().Delete(path)
		if err != nil {
			return err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return NewAPIError(resp.StatusCode(), resp.Status())
		}
		var dataResponse map[string]interface{}
		err = json.Unmarshal(resp.Body(), &dataResponse)
		if err != nil {
			return err
		}
		log.Debug("Data response: ", dataResponse)

		return nil
	}
}

// newKibanaSavedObjectExportFunc permit to export Kibana object
func newKibanaSavedObjectExportFunc(c *resty.Client) KibanaSavedObjectExport {
	return func(objectTypes []string, objects []map[string]string, deepReference bool, kibanaSpace string) ([]byte, error) {

		log.Debug("ObjectTypes: ", objectTypes)
		log.Debug("Objects: ", objects)
		log.Debug("DeepReference: ", deepReference)
		log.Debug("KibanaSpace: ", kibanaSpace)

		payload := make(map[string]interface{})
		payload["excludeExportDetails"] = true

		if len(objectTypes) > 0 {
			payload["type"] = objectTypes
		}
		if len(objects) > 0 {
			payload["objects"] = objects
		}
		payload["includeReferencesDeep"] = deepReference
		log.Debug("Payload: ", payload)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/_export", basePathKibanaSavedObject)
		} else {
			path = fmt.Sprintf("/s/%s%s/_export", kibanaSpace, basePathKibanaSavedObject)
		}
		log.Debugf("URL to export object: %s", path)

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		resp, err := c.R().SetBody(jsonData).Post(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}

		data := resp.Body()
		log.Debug("Data response: ", data)

		return data, nil

	}
}

// newKibanaSavedObjectImportFunc permit to import Kibana object
func newKibanaSavedObjectImportFunc(c *resty.Client) KibanaSavedObjectImport {
	return func(data []byte, overwrite bool, kibanaSpace string) (map[string]interface{}, error) {

		if len(data) == 0 {
			return nil, NewAPIError(600, "You must provide data parameters")
		}

		log.Debug("Data: ", data)
		log.Debug("Overwrite: ", overwrite)
		log.Debug("kibanaSpace: ", kibanaSpace)

		var path string
		if kibanaSpace == "" || kibanaSpace == "default" {
			path = fmt.Sprintf("%s/_import", basePathKibanaSavedObject)
		} else {
			path = fmt.Sprintf("/s/%s%s/_import", kibanaSpace, basePathKibanaSavedObject)
		}
		log.Debugf("URL to export object: %s", path)

		resp, err := c.R().
			SetQueryString(fmt.Sprintf("overwrite=%t", overwrite)).
			SetFileReader("file", "file.ndjson", bytes.NewReader(data)).
			Post(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		var dataResponse map[string]interface{}
		err = json.Unmarshal(resp.Body(), &dataResponse)
		if err != nil {
			return nil, err
		}
		log.Debug("Data response: ", dataResponse)

		return dataResponse, nil

	}
}
