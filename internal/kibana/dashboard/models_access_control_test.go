package dashboard

import (
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestAccessControlValue_ToCreateAPI(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var m *AccessControlValue
		apiModel := m.ToCreateAPI()
		assert.Nil(t, apiModel)
	})

	t.Run("empty values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringNull(),
		}
		apiModel := m.ToCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("private"),
			Owner:      types.StringValue("user123"),
		}
		apiModel := m.ToCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Equal(t, utils.Pointer(kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode("private")), apiModel.AccessMode)
		assert.Equal(t, utils.Pointer("user123"), apiModel.Owner)
	})

	t.Run("partial values - access_mode", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("private"),
			Owner:      types.StringNull(),
		}
		apiModel := m.ToCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Equal(t, utils.Pointer(kbapi.PostDashboardsJSONBodyDataAccessControlAccessMode("private")), apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})

	t.Run("partial values - owner", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringValue("user123"),
		}
		apiModel := m.ToCreateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		assert.Equal(t, utils.Pointer("user123"), apiModel.Owner)
	})
}

func TestAccessControlValue_ToUpdateAPI(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var m *AccessControlValue
		apiModel := m.ToUpdateAPI()
		assert.Nil(t, apiModel)
	})

	t.Run("filled values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringValue("public"),
			Owner:      types.StringValue("admin"),
		}
		apiModel := m.ToUpdateAPI()
		assert.NotNil(t, apiModel)
		assert.Equal(t, utils.Pointer(kbapi.PutDashboardsIdJSONBodyDataAccessControlAccessMode("public")), apiModel.AccessMode)
		assert.Equal(t, utils.Pointer("admin"), apiModel.Owner)
	})

	t.Run("empty values", func(t *testing.T) {
		m := &AccessControlValue{
			AccessMode: types.StringNull(),
			Owner:      types.StringNull(),
		}
		apiModel := m.ToUpdateAPI()
		assert.NotNil(t, apiModel)
		assert.Nil(t, apiModel.AccessMode)
		assert.Nil(t, apiModel.Owner)
	})
}

func TestNewAccessControlFromAPI(t *testing.T) {
	t.Run("nil inputs", func(t *testing.T) {
		val := newAccessControlFromAPI(nil, nil)
		assert.Nil(t, val)
	})

	t.Run("filled inputs", func(t *testing.T) {
		accessMode := "private"
		owner := "user1"
		val := newAccessControlFromAPI(&accessMode, &owner)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("private"), val.AccessMode)
		assert.Equal(t, types.StringValue("user1"), val.Owner)
	})

	t.Run("partial inputs - access_mode", func(t *testing.T) {
		accessMode := "private"
		val := newAccessControlFromAPI(&accessMode, nil)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringValue("private"), val.AccessMode)
		assert.Equal(t, types.StringNull(), val.Owner)
	})

	t.Run("partial inputs - owner", func(t *testing.T) {
		owner := "user1"
		val := newAccessControlFromAPI(nil, &owner)
		assert.NotNil(t, val)
		assert.Equal(t, types.StringNull(), val.AccessMode)
		assert.Equal(t, types.StringValue("user1"), val.Owner)
	})
}
