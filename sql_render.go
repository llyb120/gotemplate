package gotemplate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"gitee.com/llyb120/goscript"
)

type SqlRender struct {
	engine *TemplateEngine
	sqlMap *sqlBlocks
}

type sqlBlocks struct {
	sync.RWMutex
	blocks map[string]map[string]string
}

type ScanHandler func(fileName string, content string) error

func (t *SqlRender) Scan(scanFn func(handler ScanHandler) error) error {
	return scanFn(t.handleSingleFile)
}

func (t *SqlRender) handleSingleFile(fileName string, content string) error {
	// 获得md的一级标题
	re := regexp.MustCompile("(?is)#(.*?)\n")
	matches := re.FindAllStringSubmatch(content, 1)
	if len(matches) == 0 {
		// 没有标题的直接忽略
		return nil
		//return errors.New("no title found")
	}
	title := strings.TrimSpace(matches[0][1])
	// 获得二级标题指向的sql
	re = regexp.MustCompile("(?is)##(.*?)\n.*?```sql(?:.*?)*\n(.*?)```")
	matches = re.FindAllStringSubmatch(content, -1)
	t.sqlMap.Lock()
	defer t.sqlMap.Unlock()
	if _, ok := t.sqlMap.blocks[title]; !ok {
		t.sqlMap.blocks[title] = map[string]string{}
	}
	for _, match := range matches {
		subTitle := strings.TrimSpace(match[1])
		sql := t.handleCommand(strings.TrimSpace(match[2]))
		t.sqlMap.blocks[title][subTitle] = sql
	}
	return nil
}

func (t *SqlRender) handleCommand(sql string) string {
	// prefix command
	// prefix := `(?:\b(val|each)\b)?`
	// 特殊空格全部转为空格
	spaceRegex := regexp.MustCompile(`\t|\r`)
	sql = spaceRegex.ReplaceAllString(sql, " ")
	command := regexp.MustCompile(`^(.*?)\s*(\bby\b.*?)?(\bwhen\b.*?)?$`)
	re := regexp.MustCompile(`(?m)(.*?)--#\s*([^\n]+)`)
	return re.ReplaceAllStringFunc(sql, func(s string) string {
		var pre, middle, post string
		// 指令语句
		matches := re.FindAllStringSubmatch(s, 1)
		// 这里不用判断，一定可以匹配到
		middle = matches[0][1] + " \n"
		parts := strings.Split(matches[0][2], "$$")
		for _, part := range parts {
			part := strings.TrimSpace(part)
			// 普通指令
			commandSubMatch := command.FindAllStringSubmatch(part, 1)
			// 如果使用了when
			if commandSubMatch[0][3] != "" && strings.HasPrefix(commandSubMatch[0][3], "when") {
				// 使用when
				conditionExpr := commandSubMatch[0][3][4:]
				//fmt.Println(conditionExpr)
				pre += "{{ if exist(" + conditionExpr + ") }} \n"
				post += "{{ end }} \n"
			}
			// 处理第一个指令
			var mainCommand string
			if commandSubMatch[0][1] != "" {
				// 如果以？结尾
				commandSubMatch[0][1] = strings.TrimSpace(commandSubMatch[0][1])
				if strings.HasSuffix(commandSubMatch[0][1], "?") {
					pre += "{{ if exist(" + commandSubMatch[0][1][:len(commandSubMatch[0][1])-1] + ") }} \n"
					post += "{{ end }} \n"
					commandSubMatch[0][1] = commandSubMatch[0][1][:len(commandSubMatch[0][1])-1]
				}

				if strings.HasPrefix(commandSubMatch[0][1], "val ") {
					mainCommand = "{{ val(" + commandSubMatch[0][1][4:] + ") }}"
				} else if strings.HasPrefix(commandSubMatch[0][1], "each ") {
					mainCommand = "{{ each(" + commandSubMatch[0][1][5:] + ") }}"
				}

				if mainCommand == "" {
					mainCommand = "{{" + commandSubMatch[0][1] + "}}"
				}
			}
			// 使用by的情况
			if commandSubMatch[0][2] != "" {
				if strings.HasPrefix(commandSubMatch[0][2], "by") {
					willBeReplaced := strings.TrimSpace(commandSubMatch[0][2][2:])
					middle = strings.ReplaceAll(middle, willBeReplaced, mainCommand)
				}
			}
		}
		// 特殊指令语句

		// 一般语句

		fmt.Println(matches)
		return pre + middle + post
	})
}

func (t *SqlRender) registerStandardFunctions() map[string]any {
	return (map[string]any{
		"val": func(args ...interface{}) interface{} {
			if len(args) != 1 {
				return nil
			}
			return args[0]
		},
		"each": func(args ...interface{}) interface{} {
			if len(args) != 1 {
				return nil
			}
			return args[0]
		},
		"exist": func(arg any) bool {
			if arg == nil || arg == goscript.Undefined {
				return false
			}
			if reflect.TypeOf(arg).Kind() == reflect.Map || reflect.TypeOf(arg).Kind() == reflect.Slice {
				return reflect.ValueOf(arg).Len() > 0
			}
			return true
		},
	})
}

func (t *SqlRender) GetSql(title, subTitle string, data any) (string, error) {
	var sql = (func() string {
		t.sqlMap.RLock()
		defer t.sqlMap.RUnlock()
		if blocks, ok := t.sqlMap.blocks[title]; ok {
			if subTitle, ok := blocks[subTitle]; ok {
				return subTitle
			}
		}
		return ""
	})()
	if sql == "" {
		return "", errors.New("sql not found")
	}
	return t.engine.Render(sql, data)
}

func NewSqlRender() *SqlRender {
	sqlRender := &SqlRender{
		// engine: NewTemplateEngine(),
		sqlMap: &sqlBlocks{
			blocks: map[string]map[string]string{},
		},
	}
	sqlRender.engine = NewTemplateEngine(sqlRender.registerStandardFunctions())
	return sqlRender
}
