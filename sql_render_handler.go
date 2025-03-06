package gotemplate

type SqlRenderPhase string

var (
	ON_SLOT_RENDER SqlRenderPhase = "ON_SLOT_RENDER"
	ON_REDO_RENDER SqlRenderPhase = "ON_REDO_RENDER"
)

type SqlHandlerContext struct {
	Name    string
	Context any
}

type SqlRenderHandler func(phase SqlRenderPhase, context SqlHandlerContext, sql *string, args *[]any) error
