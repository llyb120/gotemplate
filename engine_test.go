package gotemplate

import (
	"testing"
)

func main() {
	engine := NewTemplateEngine()
	// var g errgroup.Group
	// engine.Scan(func(handler ScanHandler) {
	// 	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
	// 		if err != nil {
	// 			return err
	// 		}
	// 		if info.IsDir() {
	// 			return nil
	// 		}
	// 		if !strings.HasSuffix(path, ".md") {
	// 			return nil
	// 		}
	// 		g.Go(func() error {
	// 			content, err := os.ReadFile(path)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			handler(path, string(content))
	// 			return nil
	// 		})
	// 		return nil
	// 	})
	// })
	// if err := g.Wait(); err != nil {
	// 	fmt.Println(err)
	// }

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
