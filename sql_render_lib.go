package gotemplate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gitee.com/llyb120/goscript"
)

func (t *SqlRender) lib() map[string]any {
	return map[string]any{
		"val": func(arg any) interface{} {
			ctx := t.sqlContext.GetContext()
			ctx.params = append(ctx.params, arg)
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
				ctx.params = append(ctx.params, value.Interface())
				str += "?,"
			}
			str = strings.TrimSuffix(str, ",")
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
		"use": func(alias string, template string) string {
			ctx := t.sqlContext.GetContext()
			ctx.currentUseScope = alias
			defer func() {
				ctx.currentUseScope = "default"
			}()
			arr := strings.Split(template, ".")
			var main, sub string
			if len(arr) == 1 {
				main = ctx.fromTitle
				sub = arr[0]
			} else {
				main = arr[0]
				sub = arr[1]
			}
			sql := t.getSql(main, sub)
			if sql == "" {
				fmt.Printf("warn: 没有找到模板 %s\n", template)
				return ""
			}
			res, err := t.engine.doRender(ctx.inter, sql)
			if err != nil {
				fmt.Println("warn: use template error", err)
				return ""
			}
			return res
		},
		"hook": func(name string, content string) string {
			ctx := t.sqlContext.GetContext()
			re := regexp.MustCompile(`@\{|@\}`)
			ctx.hooks[name] = re.ReplaceAllStringFunc(content, func(s string) string {
				if s == `@{` {
					return `{{`
				} else if s == `@}` {
					return `}}`
				}
				return s
			})
			return ""
		},
		"slot": func(name string, self string) string {
			ctx := t.sqlContext.GetContext()
			var code string
			var ok bool
			if code, ok = ctx.hooks[ctx.currentUseScope+"."+name]; !ok {
				// 对自身进行转义
				re := regexp.MustCompile(`@\{|@\}`)
				code = re.ReplaceAllStringFunc(self, func(s string) string {
					if s == `@{` {
						return `{{`
					} else if s == `@}` {
						return `}}`
					}
					return s
				})
			}
			res, err := t.engine.doRender(ctx.inter, code)
			if err != nil {
				fmt.Println("warn: use template error", err)
				return ""
			}
			return res
		},
	}
}
