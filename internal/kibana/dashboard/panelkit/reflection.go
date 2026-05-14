package panelkit

import (
	"fmt"
	"reflect"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/dashboard/models"
)

var panelModelTfsdkFieldIndex map[string]int

func init() {
	t := reflect.TypeOf(models.PanelModel{})
	panelModelTfsdkFieldIndex = make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("tfsdk")
		if tag == "" {
			continue
		}
		panelModelTfsdkFieldIndex[tag] = i
	}
}

// HasPanelConfigBlock reports whether blockName is a tfsdk tag on models.PanelModel.
func HasPanelConfigBlock(blockName string) bool {
	_, ok := panelModelTfsdkFieldIndex[blockName]
	return ok
}

// MustPanelConfigBlockTagged panics if blockName is not a tfsdk tag on models.PanelModel.
func MustPanelConfigBlockTagged(blockName string) {
	if !HasPanelConfigBlock(blockName) {
		panic(fmt.Sprintf("dashboard panel registry: unknown panel config block %q — no PanelModel field with matching tfsdk tag", blockName))
	}
}

// HasConfig reports whether pm has non-nil typed state for blockName (e.g. "slo_burn_rate_config").
func HasConfig(pm *models.PanelModel, blockName string) bool {
	if pm == nil {
		return false
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		return false
	}
	rv := reflect.ValueOf(pm).Elem().Field(idx)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Interface, reflect.Map:
		return !rv.IsNil()
	default:
		return !rv.IsZero()
	}
}

// ClearConfig zeroes pm's typed field for blockName.
func ClearConfig(pm *models.PanelModel, blockName string) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		return
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() {
		return
	}
	fv.Set(reflect.Zero(fv.Type()))
}

// SetConfig assigns cfg to pm's typed field keyed by blockName. cfg must be assignable to that field type.
// A nil cfg clears the field when the field is pointer-like or otherwise zeroable.
func SetConfig(pm *models.PanelModel, blockName string, cfg any) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		panic(fmt.Sprintf("panelkit.SetConfig: unknown block %q", blockName))
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() {
		return
	}
	if cfg == nil {
		fv.Set(reflect.Zero(fv.Type()))
		return
	}
	cv := reflect.ValueOf(cfg)
	if cv.Type().AssignableTo(fv.Type()) {
		fv.Set(cv)
		return
	}
	if cv.Type().ConvertibleTo(fv.Type()) {
		fv.Set(cv.Convert(fv.Type()))
		return
	}
	panic(fmt.Sprintf("panelkit.SetConfig: value type %s not assignable to field %s (%s)", cv.Type(), blockName, fv.Type()))
}

// EnsureMutableTypedConfig allocates *T for pm's pointer-backed config block at blockName when it is nil.
func EnsureMutableTypedConfig(pm *models.PanelModel, blockName string) {
	if pm == nil {
		return
	}
	idx, ok := panelModelTfsdkFieldIndex[blockName]
	if !ok {
		return
	}
	fv := reflect.ValueOf(pm).Elem().Field(idx)
	if !fv.CanSet() || fv.Kind() != reflect.Pointer {
		return
	}
	if fv.IsNil() {
		fv.Set(reflect.New(fv.Type().Elem()))
	}
}
