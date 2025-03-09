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
			if len(ctx.inters) == 0 {
				ctx.err = fmt.Errorf("没有找到模板 %s %s", main, sub)
				return ""
			}
			var shouldUseCurrentContext = false
			if useCurrentContext := params["context"]; useCurrentContext == "current" {
				shouldUseCurrentContext = true
			}
			var inter = ctx.inters[len(ctx.inters)-1]
			if !shouldUseCurrentContext {
				inter = inter.Fork()
				ctx.inters = append(ctx.inters, inter)
				// prepare hooks
				ctx.hooks = append(ctx.hooks, make(map[string]string))
				// prepare slots
				ctx.slotHistories = append(ctx.slotHistories, make(map[string]string))
			}
			for k, v := range hookContext {
				// hook := v.(int)
				// decodeCode(&hook)
				ctx.hooks[len(ctx.hooks)-1][k] = t.sqlMap.constants[v.(int)]
			}
			defer func() {
				if !shouldUseCurrentContext {
					if len(ctx.hooks) == 0 || len(ctx.slotHistories) == 0 || len(ctx.inters) == 0 {
						ctx.err = fmt.Errorf("hook的作用域stack被错误弹出")
					} else {
						ctx.hooks = ctx.hooks[:len(ctx.hooks)-1]
						ctx.inters = ctx.inters[:len(ctx.inters)-1]
						ctx.slotHistories = ctx.slotHistories[:len(ctx.slotHistories)-1]
					}
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
				// 这里是有可能的，因为最外层的没有对应的hook作用域
				code = t.sqlMap.constants[selfIndex]
			} else if code, ok = ctx.hooks[len(ctx.hooks)-1][name]; !ok {
				// 对自身进行转义
				// decodeCode(&self)
				code = t.sqlMap.constants[selfIndex]
			}
			if strings.TrimSpace(code) == "" {
				return ""
			}
			if len(ctx.inters) == 0 || len(ctx.slotHistories) == 0 {
				ctx.err = fmt.Errorf("slot的作用域stack被错误弹出")
				return ""
			}
			ctx.slotHistories[len(ctx.slotHistories)-1][name] = code
			res, err := t.engine.doRender(ctx.inters[len(ctx.inters)-1], code)
			if err != nil {
				ctx.err = err
				return ""
			}
			if err := t.handlePhase(ctx, ON_SLOT_RENDER, SqlHandlerContext{
				Name:    name,
				Context: ctx.inters[len(ctx.inters)-1].GetGlobal(),
			}, &res, &ctx.params); err != nil {
				ctx.err = err
				return ""
			}
			return res
		},
		"redo": func(name string) string {
			ctx := t.sqlContext.GetContext()
			if len(ctx.slotHistories) == 0 {
				fmt.Printf("warn: redo无法找到对应的slot %s \n", name)
				return ""
			}
			var code string
			var ok bool
			if code, ok = ctx.slotHistories[len(ctx.slotHistories)-1][name]; !ok {
				fmt.Printf("warn: redo无法找到对应的slot %s \n", name)
				return ""
			}
			res, err := t.engine.doRender(ctx.inters[len(ctx.inters)-1], code)
			if err != nil {
				ctx.err = err
				return ""
			}
			if err := t.handlePhase(ctx, ON_REDO_RENDER, SqlHandlerContext{
				Name:    name,
				Context: ctx.inters[len(ctx.inters)-1].GetGlobal(),
			}, &res, &ctx.params); err != nil {
				ctx.err = err
				return ""
			}
			return res
		},
		"trim": func(target, safe string, contentIndex int) string {
			ctx := t.sqlContext.GetContext()
			content := t.sqlMap.constants[contentIndex]
			// decodeCode(&content)
			if len(ctx.inters) == 0 {
				ctx.err = fmt.Errorf("trim的作用域stack被错误弹出")
				return ""
			}
			res, err := t.engine.doRender(ctx.inters[len(ctx.inters)-1], content)
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

func (t *SqlRender) handlePhase(ctx *sqlContextItem, phase SqlRenderPhase, context SqlHandlerContext, sql *string, args *[]any) error {
	for _, handler := range ctx.handlers {
		if err := handler(phase, context, sql, args); err != nil {
			return err
		}
	}
	return nil
}
