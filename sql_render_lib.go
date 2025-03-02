package gotemplate

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
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
			ctx := t.sqlContext.GetContext()
			for i := 0; i < iter.Len(); i++ {
				value := iter.Index(i)
				ctx.params = append(ctx.params, value.Interface())
				str += "?,"
			}
			str = strings.TrimSuffix(str, ",")
			if str == "" {
				fmt.Println("warn: each 没有找到数据")
				ctx.params = append(ctx.params, "__UNDEFINED__")
				return "?"
			}
			return str
		},
		"use": func(alias, main, sub string, params map[string]any) string {
			ctx := t.sqlContext.GetContext()
			ctx.currentUseScope = alias
			defer func() {
				ctx.currentUseScope = "default"
			}()
			if main == "" {
				main = ctx.title
			}
			if sub == "self" {
				sub = ctx.subTitle
			}
			sql := t.getSql(main, sub)
			if sql == "" {
				ctx.err = fmt.Errorf("没有找到模板 %s %s", main, sub)
				return ""
			}
			// use 应当开启一个新的作用域
			inter := ctx.inter.Fork()
			for k, v := range params {
				inter.Set(k, v)
			}
			res, err := t.engine.doRender(inter, sql)
			if err != nil {
				ctx.err = err
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
				decodeCode(&self)
				code = self
			}
			res, err := t.engine.doRender(ctx.inter, code)
			if err != nil {
				ctx.err = err
				return ""
			}
			return res
		},
		"trim": func(target, safe, content string) string {
			ctx := t.sqlContext.GetContext()
			decodeCode(&content)
			res, err := t.engine.doRender(ctx.inter, content)
			if err != nil {
				ctx.err = err
				return ""
			}
			res = strings.TrimSpace(res)
			res = strings.TrimPrefix(res, target)
			res = strings.TrimSuffix(res, target)
			if strings.TrimSpace(res) == "" {
				return fmt.Sprintf("\n %s \n", safe)
			}
			// 最后要补上空格以防报错
			return fmt.Sprintf("\n %s \n", res)
		},
	}
}
