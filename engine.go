package gotemplate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/llyb120/goscript"
)

type TemplateEngine struct {
	interpreter *goscript.Interpreter
	parsedCache *parsedCache
}

func (t *TemplateEngine) Render(template string, data any) (string, error) {
	// 模板预处理
	inter, err := t.prepareRender(data)
	if err != nil {
		return "", err
	}
	return t.doRender(inter, template)
}

func (t *TemplateEngine) prepareRender(data any) (*goscript.Interpreter, error) {
	inter := t.interpreter.Fork()
	inter.SetGlobal(data)
	return inter, nil
}

func (t *TemplateEngine) doRender(inter *goscript.Interpreter, template string) (string, error) {
	code := t.parsedCache.GetIfNotExist(template, func() string {
		return t.preHandle(template)
	})
	if code == "" {
		return "", errors.New("template is not parsed")
	}
	// string package
	result, err := inter.Interpret(code)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", errors.New("result is nil")
	}
	if resultStr, ok := result.(string); ok {
		return resultStr, nil
	}
	return "", errors.New("result is not a string")
}

func (t *TemplateEngine) preHandle(content string) string {
	re := regexp.MustCompile(`(?s)\{\{(.*?)\}\}`)
	// 0 - 1 out start end
	// 2 - 3 command start end
	ctrlStmtReg := regexp.MustCompile(`^(\bif\b|\bfor\b|\belse\b|\belse\b\s+\bif\b)`)
	indexes := re.FindAllStringSubmatchIndex(content, -1)
	// ss := re.FindAllStringSubmatch(content, -1)
	var builder strings.Builder
	builder.WriteString("var __code__ strings.Builder \n")
	var pos = 0
	for _, index := range indexes {
		staticContent := content[pos:index[0]]
		// 对staticContent进行转义
		escapeBacktick(&staticContent)
		builder.WriteString(fmt.Sprintf("__code__.WriteString(`%s`) \n", staticContent))
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
				if strings.HasPrefix(command, "for") {
					builder.WriteString(fmt.Sprintf("%s {\n", command))
				} else {
					// if or else if
					if strings.HasPrefix(command, "if") {
						command = strings.TrimPrefix(command, "if")
						builder.WriteString(fmt.Sprintf("if %s {\n", command))
					} else {
						command = strings.TrimPrefix(command, "else")
						command = strings.TrimLeft(command, " ")
						command = strings.TrimPrefix(command, "if")
						builder.WriteString(fmt.Sprintf("else if %s {\n", command))
					}
				}
			}
		} else if strings.HasPrefix(command, "end") {
			builder.WriteString("} \n")
		} else {
			builder.WriteString(fmt.Sprintf("__code__.WriteString(fmt.Sprintf(\"%%v\",str(%s))) \n", content[index[2]:index[3]]))
		}
		builder.WriteString(" \n")
		// builder.WriteString(fmt.Sprintf(`builder.WriteString("%s")`, content[index[2]:index[3]]))
		pos = index[1]
	}
	// 如果还有尾部
	if pos < len(content) {
		staticContent := content[pos:]
		// 对staticContent进行转义
		escapeBacktick(&staticContent)
		builder.WriteString(fmt.Sprintf("__code__.WriteString(`%s`) \n", staticContent))
	}
	builder.WriteString("return __code__.String() \n")
	return builder.String()
}

func NewTemplateEngine(scope map[string]any) *TemplateEngine {
	engine := &TemplateEngine{
		interpreter: goscript.NewInterpreter(),
		parsedCache: &parsedCache{
			cache: make(map[string]string),
		},
	}
	for k, v := range engine.lib() {
		engine.interpreter.Set(k, v)
	}
	for key, value := range scope {
		engine.interpreter.Set(key, value)
	}
	return engine
}
