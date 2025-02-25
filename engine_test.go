package gotemplate

import (
	"fmt"
	"testing"
)

func main() {
	engine := NewTemplateEngine(nil)

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

	res, _ := engine.Render(template, map[string]interface{}{
		"a": 1,
		"b": 2,
	})
	fmt.Println(res)
}

func TestTemplateEngine_Render(t *testing.T) {
	main()
}
