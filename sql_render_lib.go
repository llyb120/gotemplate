package gotemplate

import (
	"reflect"
	"strings"

	"gitee.com/llyb120/goscript"
)

func (t *SqlRender) lib() map[string]any {
	return map[string]any{
		"val": func(arg any) interface{} {
			ctx := t.sqlContext.GetContext()
			(*ctx) = append(*ctx, arg)
			return "?"
		},
		"each": func(arg any) string {
			// 必须是一个切片，通过反射进行判断
			if reflect.TypeOf(arg).Kind() != reflect.Slice {
				return ""
			}
			// 循环
			iter := reflect.ValueOf(arg)
			str := ""
			for i := 0; i < iter.Len(); i++ {
				value := iter.Index(i)
				ctx := t.sqlContext.GetContext()
				(*ctx) = append(*ctx, value.Interface())
				str += "?,"
			}
			if strings.HasSuffix(str, ",") {
				str = str[:len(str)-1]
			}
			if str == "" {
				return "__UNDEFINED__"
			}
			return str
		},
		"exist": func(arg any) bool {
			if arg == nil || arg == goscript.Undefined {
				return false
			}
			if reflect.TypeOf(arg).Kind() == reflect.Map || reflect.TypeOf(arg).Kind() == reflect.Slice {
				return reflect.ValueOf(arg).Len() > 0
			}
			return true
		},
	}
}
