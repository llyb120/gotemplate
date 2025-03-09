package gotemplate

import (
	"fmt"
)

func (t *TemplateEngine) lib() map[string]any {
	return map[string]any{
		"str": _str,
		// "when": _when,
	}
}

func _str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// func _when(v any) bool {
// 	if v == nil || v == goscript.Undefined {
// 		return false
// 	}
// 	if v == false || v == 0 || v == "" {
// 		return false
// 	}
// 	if reflect.TypeOf(v).Kind() == reflect.Map || reflect.TypeOf(v).Kind() == reflect.Slice {
// 		return reflect.ValueOf(v).Len() > 0
// 	}
// 	return true
// }
