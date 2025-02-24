package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitee.com/llyb120/goscript"
	"golang.org/x/sync/errgroup"
)

type TemplateEngine struct {
	interpreter *goscript.Interpreter
}

// func (t *TemplateEngine) Render(template string, data any) (string, error) {
// 	return t.interpreter.Interpret(template, data)
// }

type ScanHandler func(fileName string, content string)

func (t *TemplateEngine) Scan(scanFn func(handler ScanHandler)) error {
	scanFn(t.handleSingleFile)
	return nil
}

func (t *TemplateEngine) handleSingleFile(fileName string, content string) {
	re := regexp.MustCompile("(?is)##(.*?)\n.*?```sql(.*?)```")
	matches := re.FindAllStringSubmatch(content, -1)
	blocks := map[string]string{}
	for _, match := range matches {
		blocks[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
		// fmt.Println(match[1])
		// fmt.Println(match[2])
	}
	// re.ReplaceAllStringFunc(content, func(match string) string {
	// 	fmt.Println(match)
	// 	return ""
	// })
	fmt.Println(blocks)

}

func (t *TemplateEngine) Render(template string, data any) (string, error) {
	// 模板预处理
	inter := goscript.NewInterpreter()
	inter.BindGlobalObject(data)
	return t.preHandle(template), nil
	// return inter.Interpret(template)
}

func (t *TemplateEngine) preHandle(content string) string {
	re := regexp.MustCompile(`(?s)\{\{(.*?)\}\}`)
	indexes := re.FindAllStringSubmatchIndex(content, -1)
	ss := re.FindAllStringSubmatch(content, -1)
	fmt.Println(indexes, ss)
	// blocks := map[string]string{}
	// for _, index := range indexes {
	// 	blocks[strings.TrimSpace(content[index[2]:index[3]])] = strings.TrimSpace(content[index[4]:index[5]])
	// }
	// return blocks
	// 	blocks[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
	// }
	return content
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		interpreter: goscript.NewInterpreter(),
	}
}

func main() {
	dir := "D:\\project\\intelligence-pc-backend\\services\\tables"
	engine := NewTemplateEngine()
	var g errgroup.Group
	engine.Scan(func(handler ScanHandler) {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			g.Go(func() error {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				handler(path, string(content))
				return nil
			})
			return nil
		})
	})
	if err := g.Wait(); err != nil {
		fmt.Println(err)
	}

	template := `
		{{a}} holy
		{{ if b > 0 }}
		 123321
		{{end}}
	`
	engine.Render(template, map[string]interface{}{
		"a": 1,
		"b": 2,
	})
}
