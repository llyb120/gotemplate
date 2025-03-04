package gotemplate

type SqlRenderHandlerPhase string

var (
	ON_SLOT_RENDER SqlRenderHandlerPhase = "ON_SLOT_RENDER"
	ON_REDO_RENDER SqlRenderHandlerPhase = "ON_REDO_RENDER"
)

type SqlRenderHandler func(phase SqlRenderHandlerPhase, context map[string]any, sql *string, args *[]any) error
