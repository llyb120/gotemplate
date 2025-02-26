package gotemplate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type SqlRender struct {
	engine     *TemplateEngine
	sqlMap     *sqlBlocks
	sqlContext *sqlContext
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
		// return nil
		return fmt.Errorf("%s 没有找到标题", fileName)
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
		sql := strings.TrimSpace(match[2])
		t.handleCommand(&sql)
		if err := t.handleSpecialCommand(&sql); err != nil {
			return err
		}
		t.sqlMap.blocks[title][subTitle] = sql
	}
	return nil
}

// 处理特殊的指令
func (t *SqlRender) handleSpecialCommand(sql *string) error {
	re := regexp.MustCompile(`(?im)^\s*--#\s*\b(use|hook|slot|end|for|if|else)\b(.*?)$`)
	matches := re.FindAllStringSubmatchIndex(*sql, -1)
	contents := re.FindAllStringSubmatch(*sql, -1)
	_ = contents

	// 记录所有hook块
	var hookBlocks []string

	// 扫描所有指令
	var builder strings.Builder
	pos := 0
	count := 0
	for i := 0; i < len(matches); i++ {
		// 写入pos到match[0]之间的内容
		builder.WriteString((*sql)[pos:matches[i][0]])
		pos = matches[i][1]
		cmdType := (*sql)[matches[i][2]:matches[i][3]]
		cmdArgs := strings.TrimSpace((*sql)[matches[i][4]:matches[i][5]])
		switch cmdType {
		case "hook", "slot":
			// 向下找到结束标识
			count = 0
			for j := i + 1; j < len(matches); j++ {
				endCmdType := (*sql)[matches[j][2]:matches[j][3]]
				if endCmdType == "end" {
					if count == 0 {
						content := (*sql)[matches[i][1]:matches[j][0]]
						if err := t.handleSpecialCommand(&content); err != nil {
							return err
						}
						// 对所有指令进行转义，否则会出错
						encodeCode(&content)
						if cmdType == "hook" {
							arr := strings.Split(cmdArgs, ".")
							var hookName string
							if len(arr) == 1 {
								hookName = "default." + arr[0]
							} else {
								hookName = arr[0] + "." + arr[1]
							}
							hookBlocks = append(hookBlocks, fmt.Sprintf("{{ \n hook(`%s`, `%s`) \n}}\n", hookName, content))
						} else if cmdType == "slot" {
							builder.WriteString(fmt.Sprintf("{{ \n __code__.WriteString(slot(`%s`, `%s`) ) \n}}\n", cmdArgs, content))
						}
						i = j
						pos = matches[j][1]
						break
					} else {
						count--
					}
				} else {
					count++
				}
			}
		case "use":
			// 如果指定了别名
			var alias, main, sub string
			index := strings.Index(cmdArgs, " as ")
			if index != -1 {
				alias = strings.TrimSpace(cmdArgs[index+4:])
				cmdArgs = strings.TrimSpace(cmdArgs[:index])
			} else {
				alias = "default"
			}
			arr := strings.Split(cmdArgs, ".")
			if len(arr) == 1 {
				// 本文件
				main = ""
				sub = arr[0]
			} else {
				// 其他文件
				main = arr[0]
				sub = arr[1]
			}
			builder.WriteString(fmt.Sprintf("{{ use(`%s`,`%s`,`%s`) }} \n", alias, main, sub))
		default:
			builder.WriteString(fmt.Sprintf("{{ %s %s }} \n", cmdType, cmdArgs))
		}
	}
	if pos < len(*sql) {
		builder.WriteString((*sql)[pos:])
	}
	if count != 0 {
		return errors.New("有未闭合的指令")
	}
	*sql = strings.Join(hookBlocks, "\n") + builder.String()
	return nil
}

func (t *SqlRender) handleCommand(sql *string) {
	// prefix command
	// prefix := `(?:\b(val|each)\b)?`
	// 特殊空格全部转为空格
	spaceRegex := regexp.MustCompile(`\t|\r`)
	*sql = spaceRegex.ReplaceAllString(*sql, " ")
	command := regexp.MustCompile(`^(.*?)\s*(\bby\b.*?)?(\bwhen\b.*?)?$`)
	re := regexp.MustCompile(`(?m)(.*?)--#\s*([^\n]+)`)
	eatRe := regexp.MustCompile(`('.*?'|".*?"|[\d\.]+)\s*,?\s*$`)
	*sql = re.ReplaceAllStringFunc(*sql, func(s string) string {
		var pre, middle, post string
		// 指令语句
		matches := re.FindAllStringSubmatch(s, 1)
		// 这里不用判断，一定可以匹配到
		middle = matches[0][1]
		if middle == "" {
			// 行指令不该由你处理
			return s
		}
		// 只处理尾指令
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
			var eatTail bool
			if commandSubMatch[0][1] != "" {
				// 如果以？结尾
				commandSubMatch[0][1] = strings.TrimSpace(commandSubMatch[0][1])
				leftCommand := commandSubMatch[0][1]

				if strings.HasPrefix(commandSubMatch[0][1], "val ") {
					leftCommand = leftCommand[4:]
					if strings.HasSuffix(leftCommand, "?") {
						pre += "{{ if exist(" + leftCommand[:len(leftCommand)-1] + ") }} \n"
						post += "{{ end }} \n"
						// 去掉？
						leftCommand = leftCommand[:len(leftCommand)-1]
					}
					mainCommand = "{{ val(" + leftCommand + ") }}"
					eatTail = true
				} else if strings.HasPrefix(commandSubMatch[0][1], "each ") {
					leftCommand = leftCommand[5:]
					if strings.HasSuffix(leftCommand, "?") {
						pre += "{{ if exist(" + leftCommand[:len(leftCommand)-1] + ") }} \n"
						post += "{{ end }} \n"
						// 去掉？
						leftCommand = leftCommand[:len(leftCommand)-1]
					}
					mainCommand = "{{ each(" + leftCommand + ") }}"
					eatTail = true
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
			} else {
				// 否则 val 和 each 需要向前吞噬
				if eatTail {
					middle = eatRe.ReplaceAllStringFunc(middle, func(s string) string {
						s = strings.TrimSpace(s)
						if strings.HasSuffix(s, ",") {
							return mainCommand + ","
						} else {
							return mainCommand
						}
					})
				}
			}
		}

		if middle != "" {
			middle += " \n"
		}
		return pre + middle + post
	})
}

func (t *SqlRender) getSql(title, subTitle string) string {
	t.sqlMap.RLock()
	defer t.sqlMap.RUnlock()
	if blocks, ok := t.sqlMap.blocks[title]; ok {
		if subTitle, ok := blocks[subTitle]; ok {
			return subTitle
		}
	}
	return ""
}

func (t *SqlRender) GetSql(title, subTitle string, data any) (string, []any, error) {
	sql := t.getSql(title, subTitle)
	if sql == "" {
		return "", nil, errors.New("sql not found")
	}
	ctx := &sqlContextItem{
		fromTitle: title,
		params:    make([]any, 0),
		hooks:     map[string]string{},
	}
	t.sqlContext.SetContext(ctx)
	defer t.sqlContext.CleanContext()
	inter, err := t.engine.prepareRender(sql, data)
	if err != nil {
		return "", nil, err
	}
	ctx.inter = inter
	sql, err = t.engine.doRender(inter, sql)
	if err != nil {
		return "", nil, err
	}
	return sql, ctx.params, ctx.err
}

func NewSqlRender() *SqlRender {
	sqlRender := &SqlRender{
		sqlContext: &sqlContext{},
		sqlMap: &sqlBlocks{
			blocks: map[string]map[string]string{},
		},
	}
	sqlRender.engine = NewTemplateEngine(sqlRender.lib())
	return sqlRender
}
