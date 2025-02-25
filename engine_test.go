package gotemplate

import (
	"testing"
)

func main() {
	engine := NewTemplateEngine()

	template := `
	{{
		var a = 2
		fmt.Println(a)
	}}

	{{a}} holy

	{{ for i := 0; i < 10; i++ }}
		{{ if i % 2 == 0 }}
			{{i}}-{{i + 1}}
		{{ else }}
			{{i}}
		{{end}}
	{{end}}
	`

	engine.Render(template, map[string]interface{}{
		"a": 1,
		"b": 2,
	})
}

func TestTemplateEngine_Render(t *testing.T) {
	main()
}
