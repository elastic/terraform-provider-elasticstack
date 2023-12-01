package kbapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	basePathKibanaSpace = "/api/spaces" // Base URL to access on Kibana space API
)

// KibanaSpace is the Space API object
type KibanaSpace struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	DisabledFeatures []string `json:"disabledFeatures,omitempty"`
	Reserved         bool     `json:"_reserved,omitempty"`
	Initials         string   `json:"initials,omitempty"`
	Color            string   `json:"color,omitempty"`
}

// KibanaSpaces is the list of KibanaSpace object
type KibanaSpaces []KibanaSpace

// KibanaSpaceCopySavedObjectParameter is parameters to copy dashboard between spaces
type KibanaSpaceCopySavedObjectParameter struct {
	Spaces            []string                     `json:"spaces"`
	IncludeReferences bool                         `json:"includeReferences"`
	Overwrite         bool                         `json:"overwrite"`
	CreateNewCopies   bool                         `json:"createNewCopies"`
	Objects           []KibanaSpaceObjectParameter `json:"objects"`
}

// KibanaSpaceObjectParameter is Object object
type KibanaSpaceObjectParameter struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// KibanaSpaceGet permit to get space
type KibanaSpaceGet func(id string) (*KibanaSpace, error)

// KibanaSpaceList permit to get all spaces
type KibanaSpaceList func() (KibanaSpaces, error)

// KibanaSpaceCreate permit to create space
type KibanaSpaceCreate func(kibanaSpace *KibanaSpace) (*KibanaSpace, error)

// KibanaSpaceDelete permit to delete space
type KibanaSpaceDelete func(id string) error

// KibanaSpaceUpdate permit to update space
type KibanaSpaceUpdate func(kibanaSpace *KibanaSpace) (*KibanaSpace, error)

// KibanaSpaceCopySavedObjects permit to copy dashboad between space
type KibanaSpaceCopySavedObjects func(parameter *KibanaSpaceCopySavedObjectParameter, spaceOrigin string) error

// String permit to return KibanaSpace object as JSON string
func (k *KibanaSpace) String() string {
	json, _ := json.Marshal(k)
	return string(json)
}

// newKibanaSpaceGetFunc permit to get the kibana space with it id
func newKibanaSpaceGetFunc(c *resty.Client) KibanaSpaceGet {
	return func(id string) (*KibanaSpace, error) {

		if id == "" {
			return nil, NewAPIError(600, "You must provide kibana space ID")
		}
		log.Debug("ID: ", id)

		path := fmt.Sprintf("%s/space/%s", basePathKibanaSpace, id)
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
		kibanaSpace := &KibanaSpace{}
		err = json.Unmarshal(resp.Body(), kibanaSpace)
		if err != nil {
			return nil, err
		}
		log.Debug("KibanaSpace: ", kibanaSpace)

		return kibanaSpace, nil
	}

}

// newKibanaSpaceListFunc permit to get all Kibana space
func newKibanaSpaceListFunc(c *resty.Client) KibanaSpaceList {
	return func() (KibanaSpaces, error) {

		path := fmt.Sprintf("%s/space", basePathKibanaSpace)
		resp, err := c.R().Get(path)
		if err != nil {
			return nil, err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		kibanaSpaces := make(KibanaSpaces, 0, 1)
		err = json.Unmarshal(resp.Body(), &kibanaSpaces)
		if err != nil {
			return nil, err
		}
		log.Debug("KibanaSpaces: ", kibanaSpaces)

		return kibanaSpaces, nil
	}

}

// newKibanaSpaceCreateFunc permit to create new Kibana space
func newKibanaSpaceCreateFunc(c *resty.Client) KibanaSpaceCreate {
	return func(kibanaSpace *KibanaSpace) (*KibanaSpace, error) {

		if kibanaSpace == nil {
			return nil, NewAPIError(600, "You must provide kibana space object")
		}
		log.Debug("KibanaSpace: ", kibanaSpace)

		jsonData, err := json.Marshal(kibanaSpace)
		if err != nil {
			return nil, err
		}
		path := fmt.Sprintf("%s/space", basePathKibanaSpace)
		resp, err := c.R().SetBody(jsonData).Post(path)
		if err != nil {
			return nil, err
		}

		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		kibanaSpace = &KibanaSpace{}
		err = json.Unmarshal(resp.Body(), kibanaSpace)
		if err != nil {
			return nil, err
		}
		log.Debug("KibanaSpace: ", kibanaSpace)

		return kibanaSpace, nil
	}

}

// newKibanaSpaceCopySavedObjectsFunc permit to copy extings objects from user space to another userSpace
func newKibanaSpaceCopySavedObjectsFunc(c *resty.Client) KibanaSpaceCopySavedObjects {
	return func(parameter *KibanaSpaceCopySavedObjectParameter, spaceOrigin string) error {

		if parameter == nil {
			return NewAPIError(600, "You must provide parameter to copy existing objects on other user spaces")
		}
		log.Debug("Parameter: ", parameter)
		log.Debug("SpaceOrigin: ", spaceOrigin)

		var path string
		if spaceOrigin == "" || spaceOrigin == "default" {
			path = fmt.Sprintf("%s/_copy_saved_objects", basePathKibanaSpace)
		} else {
			path = fmt.Sprintf("/s/%s%s/_copy_saved_objects", spaceOrigin, basePathKibanaSpace)
		}
		jsonData, err := json.Marshal(parameter)
		if err != nil {
			return err
		}
		resp, err := c.R().SetBody(jsonData).Post(path)
		if err != nil {
			return err
		}

		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return NewAPIError(resp.StatusCode(), resp.Status())
		}
		data := make(map[string]interface{})
		err = json.Unmarshal(resp.Body(), &data)
		if err != nil {
			return err
		}
		log.Debug("Response: ", data)

		var errors []string
		for name, object := range data {
			if !object.(map[string]interface{})["success"].(bool) {
				errors = append(errors, fmt.Sprintf("Error to process user space %s", name))
			}
		}
		if len(errors) > 0 {
			return NewAPIError(500, strings.Join(errors, "\n"))
		}

		return nil
	}

}

// newKibanaSpaceDeleteFunc permit to delete the kubana space wiht it id
func newKibanaSpaceDeleteFunc(c *resty.Client) KibanaSpaceDelete {
	return func(id string) error {

		if id == "" {
			return NewAPIError(600, "You must provide kibana space ID")
		}

		log.Debug("ID: ", id)

		path := fmt.Sprintf("%s/space/%s", basePathKibanaSpace, id)
		resp, err := c.R().Delete(path)
		if err != nil {
			return err
		}
		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {

			return NewAPIError(resp.StatusCode(), resp.Status())

		}

		return nil
	}

}

// newKibanaSpaceUpdateFunc permit to update the Kibana space
func newKibanaSpaceUpdateFunc(c *resty.Client) KibanaSpaceUpdate {
	return func(kibanaSpace *KibanaSpace) (*KibanaSpace, error) {

		if kibanaSpace == nil {
			return nil, NewAPIError(600, "You must provide kibana space object")
		}
		log.Debug("KibanaSpace: ", kibanaSpace)

		jsonData, err := json.Marshal(kibanaSpace)
		if err != nil {
			return nil, err
		}
		path := fmt.Sprintf("%s/space/%s", basePathKibanaSpace, kibanaSpace.ID)
		resp, err := c.R().SetBody(jsonData).Put(path)
		if err != nil {
			return nil, err
		}

		log.Debug("Response: ", resp)
		if resp.StatusCode() >= 300 {
			return nil, NewAPIError(resp.StatusCode(), resp.Status())
		}
		kibanaSpace = &KibanaSpace{}
		err = json.Unmarshal(resp.Body(), kibanaSpace)
		if err != nil {
			return nil, err
		}
		log.Debug("KibanaSpace: ", kibanaSpace)

		return kibanaSpace, nil
	}

}
