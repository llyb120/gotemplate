package gotemplate

type SqlRenderHandlerPhase string

var (
	ON_SLOT_RENDER SqlRenderHandlerPhase = "OnSlotRender"
)

type SqlRenderHandler func(phase SqlRenderHandlerPhase, sql *string, args *[]any) error
