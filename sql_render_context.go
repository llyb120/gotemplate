package gotemplate

import (
	"sync"

	"gitee.com/llyb120/goscript"
	"github.com/petermattis/goid"
)

type sqlContext struct {
	sync.Map
}
type sqlContextItem struct {
	fromTitle       string
	params          []any
	hooks           map[string]string
	inter           *goscript.Interpreter
	currentUseScope string
}

func (ctx *sqlContext) SetContext(sqlContextItem *sqlContextItem) {
	ctx.Store(goid.Get(), sqlContextItem)
}

func (ctx *sqlContext) GetContext() *sqlContextItem {
	if ctx, ok := ctx.Load(goid.Get()); ok {
		return ctx.(*sqlContextItem)
	}
	return nil
}

func (ctx *sqlContext) CleanContext() {
	ctx.Delete(goid.Get())
}
