package gotemplate

import (
	"sync"

	"gvisor.dev/gvisor/pkg/goid"
)

type sqlContext struct {
	sync.Map
}

func (ctx *sqlContext) SetContext(params *[]any) {
	ctx.Store(goid.Get(), params)
}

func (ctx *sqlContext) GetContext() *[]any {
	if ctx, ok := ctx.Load(goid.Get()); ok {
		return ctx.(*[]any)
	}
	return nil
}

func (ctx *sqlContext) CleanContext() {
	ctx.Delete(goid.Get())
}
