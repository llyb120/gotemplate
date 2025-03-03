package gotemplate

import (
	"fmt"
	"reflect"
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
		"use": func(main, sub string, params map[string]any, hookContext map[string]any) string {
			ctx := t.sqlContext.GetContext()
			// var oldCurrentUseScope = ctx.currentUseScope
			// if oldCurrentUseScope == "" {
			// 	oldCurrentUseScope = "default"
			// }
			// ctx.currentUseScope = alias
			// defer func() {
			// 	ctx.currentUseScope = oldCurrentUseScope
			// }()
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
			ctx.inter = inter
			// prepare hooks
			ctx.hooks = append(ctx.hooks, make(map[string]string))
			for k, v := range hookContext {
				// hook := v.(int)
				// decodeCode(&hook)
				ctx.hooks[len(ctx.hooks)-1][k] = t.sqlMap.constants[v.(int)]
			}
			defer func() {
				if len(ctx.hooks) == 0 {
					ctx.err = fmt.Errorf("hook的作用域stack被错误弹出")
				} else {
					ctx.hooks = ctx.hooks[:len(ctx.hooks)-1]
				}
			}()
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
		"slot": func(name string, selfIndex int) string {
			ctx := t.sqlContext.GetContext()
			var code string
			var ok bool
			if len(ctx.hooks) == 0 {
				ctx.err = fmt.Errorf("没有找到对应的hook作用域")
				return ""
			}
			if code, ok = ctx.hooks[len(ctx.hooks)-1][name]; !ok {
				// 对自身进行转义
				// decodeCode(&self)
				code = t.sqlMap.constants[selfIndex]
			}
			res, err := t.engine.doRender(ctx.inter, code)
			if err != nil {
				ctx.err = err
				return ""
			}
			return res
		},
		"trim": func(target, safe string, contentIndex int) string {
			ctx := t.sqlContext.GetContext()
			content := t.sqlMap.constants[contentIndex]
			// decodeCode(&content)
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
