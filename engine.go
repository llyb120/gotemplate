package gotemplate

import (
	"fmt"
	"regexp"
	"strings"

	"gitee.com/llyb120/goscript"
)

type TemplateEngine struct {
	interpreter *goscript.Interpreter
	parsedCache *parsedCache
}

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
	}
}

func (t *TemplateEngine) Render(template string, data any) {
	// 模板预处理
	inter := goscript.NewInterpreter()
	inter.BindGlobalObject(data)
	code := t.parsedCache.GetIfNotExist(template, func() string {
		return t.preHandle(template)
	})
	if code == "" {
		fmt.Println("template is not parsed")
		return
	}
	// string package
	result, err := inter.Interpret(code)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)
	// return result, nil
}

func (t *TemplateEngine) preHandle(content string) string {
	re := regexp.MustCompile(`(?s)\{\{(.*?)\}\}`)
	// 0 - 1 out start end
	// 2 - 3 command start end
	ctrlStmtReg := regexp.MustCompile(`^(\bif\b|\bfor\b|\belse\b)`)
	indexes := re.FindAllStringSubmatchIndex(content, -1)
	// ss := re.FindAllStringSubmatch(content, -1)
	var builder strings.Builder
	builder.WriteString("var code strings.Builder \n")
	var pos = 0
	for _, index := range indexes {
		staticContent := content[pos:index[0]]
		// 对staticContent进行转义
		staticContent = strings.ReplaceAll(staticContent, "`", `\`+"`")
		builder.WriteString(fmt.Sprintf("code.WriteString(`%s`) \n", staticContent))
		sourceCommand := content[index[2]:index[3]]
		command := strings.TrimSpace(sourceCommand)
		if strings.Contains(sourceCommand, "\n") {
			builder.WriteString(sourceCommand)
			builder.WriteString("\n")
		} else if ctrlStmtReg.MatchString(command) {
			// 特殊处理else
			if command == "else" {
				builder.WriteString("} else {\n")
			} else {
				builder.WriteString(fmt.Sprintf("%s {\n", command))
			}
		} else if strings.HasPrefix(command, "end") {
			builder.WriteString("} \n")
		} else {
			builder.WriteString(fmt.Sprintf("code.WriteString(fmt.Sprintf(\"%%v\",%s)) \n", content[index[2]:index[3]]))
		}
		builder.WriteString(" \n")
		// builder.WriteString(fmt.Sprintf(`builder.WriteString("%s")`, content[index[2]:index[3]]))
		pos = index[1]
	}
	builder.WriteString("return code.String() \n")
	code := builder.String()
	fmt.Println(code)
	// fmt.Println(indexes, ss)
	// blocks := map[string]string{}
	// for _, index := range indexes {
	// 	blocks[strings.TrimSpace(content[index[2]:index[3]])] = strings.TrimSpace(content[index[4]:index[5]])
	// }
	// return blocks
	// 	blocks[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
	// }
	return builder.String()
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		interpreter: goscript.NewInterpreter(),
		parsedCache: NewParsedCache(),
	}
}
