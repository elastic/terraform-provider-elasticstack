package saved_object

import (
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ksoModelV0 struct {
	ID       types.String `tfsdk:"id"`
	SpaceID  types.String `tfsdk:"space_id"`
	Object   types.String `tfsdk:"object"`
	Imported types.String `tfsdk:"imported"`
	Type     types.String `tfsdk:"type"`
}

func (m *ksoModelV0) UpdateModelWithObject() error {
	var object map[string]any

	err := json.Unmarshal([]byte(m.Object.ValueString()), &object)
	if err != nil {
		return err
	}

	if objType, ok := object["type"]; ok {
		m.Type = types.StringValue(objType.(string))
	} else {
		return errors.New("missing 'type' field in JSON object")
	}

	if objId, ok := object["id"]; ok {
		m.ID = types.StringValue(objId.(string))
	} else {
		return errors.New("missing 'id' field in JSON object")
	}

	ksoRemoveUnwantedFields(object)

	imported, err := json.Marshal(object)
	if err != nil {
		return err
	}
	m.Imported = types.StringValue(string(imported))
	return nil
}

func ksoRemoveUnwantedFields(object map[string]any) {
	// remove fields carrying state
	delete(object, "created_at")
	delete(object, "created_by")
	delete(object, "updated_at")
	delete(object, "updated_by")
	delete(object, "version")
	delete(object, "migrationVersion")
	delete(object, "namespaces")
}
