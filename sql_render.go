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
	blocks map[string]string
	// 常量池
	constants map[int]string
}

// type sqlBlock struct {
// 	sql       string
// 	constants map[int]string
// }

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
	for _, match := range matches {
		subTitle := strings.TrimSpace(match[1])
		sql := strings.TrimSpace(match[2])
		t.handleCommand(&sql)
		if err := t.handleSpecialCommand(&sql, nil); err != nil {
			return err
		}
		t.sqlMap.Lock()
		t.sqlMap.blocks[title+":"+subTitle] = sql
		t.sqlMap.Unlock()
	}
	return nil
}

// 处理特殊的指令
var lineCommandRe = regexp.MustCompile(`(?im)^\s*--#\s*\b(use|hook|slot|trim|end|for|if|else|redo)\b(.*?)(\bif\b.*?)?$`)

func (t *SqlRender) handleSpecialCommand(sql *string, hookContext *string) error {
	matches := lineCommandRe.FindAllStringSubmatchIndex(*sql, -1)
	contents := lineCommandRe.FindAllStringSubmatch(*sql, -1)
	_ = contents

	// 记录所有hook块
	// var hookBlocks []string

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
		ifExpr := ""
		if matches[i][6] != -1 && matches[i][7] != -1 {
			ifExpr = strings.TrimPrefix((*sql)[matches[i][6]:matches[i][7]], "if")
		}
		switch cmdType {
		case "hook", "slot", "trim", "use":
			if cmdType == "use" {
				str := "map[string]any{"
				hookContext = &str
			}
			// 向下找到结束标识
			count = 0
			for j := i + 1; j < len(matches); j++ {
				endCmdType := (*sql)[matches[j][2]:matches[j][3]]
				if endCmdType == "end" {
					if count == 0 {
						content := (*sql)[matches[i][1]:matches[j][0]]
						if err := t.handleSpecialCommand(&content, hookContext); err != nil {
							return err
						}
						// 对所有指令进行转义，否则会出错
						// encodeCode(&content)
						// escapeBacktick(&content)
						t.sqlMap.Lock()
						constantIndex := len(t.sqlMap.constants)
						t.sqlMap.constants[constantIndex] = content
						t.sqlMap.Unlock()
						if cmdType == "hook" {
							if hookContext == nil {
								return fmt.Errorf("hook指令必须在use指令中使用")
							}
							*hookContext += fmt.Sprintf("`%s`: %d,", cmdArgs, constantIndex)
						} else if cmdType == "slot" {
							builder.WriteString(fmt.Sprintf("{{ \n __code__.WriteString(slot(`%s`, %d) ) \n}}\n", cmdArgs, constantIndex))
						} else if cmdType == "trim" {
							// 分析trim指令
							re := regexp.MustCompile(`(.*?)\s*(?:safe\s*(.*?))?$`)
							matches := re.FindAllStringSubmatch(cmdArgs, 1)
							if len(matches) == 0 {
								return fmt.Errorf("trim指令格式错误")
							}
							builder.WriteString(fmt.Sprintf("{{ \n __code__.WriteString(trim(`%s`, `%s`, %d)) \n}}\n", matches[0][1], matches[0][2], constantIndex))
						} else if cmdType == "use" {
							func() {
								p := strings.Index(cmdArgs, ":")
								var cmdParams string
								if p > -1 {
									cmdParams = cmdArgs[p+1:]
									cmdArgs = cmdArgs[:p]
								}
								res := regexp.MustCompile(`^(?:(.*?)\.)?(.*?)\s*$`).FindAllStringSubmatch(cmdArgs, 1)
								// 如果指定了别名
								var main, sub string
								main = res[0][1]
								sub = res[0][2]
								var params string = "map[string]string{"
								if p > -1 {
									re := regexp.MustCompile(`(\w+(?:\.\w+)?)\s*=\s*(?:"([^"]*)"|'([^']*)'|(\w+))`)
									res := re.FindAllStringSubmatch(cmdParams, -1)
									// params := make(map[string]string)
									for _, match := range res {
										key := match[1]
										var value string

										// Check which capture group has the value
										switch {
										case match[2] != "": // Matched double-quoted string
											value = match[2]
										case match[3] != "": // Matched single-quoted string
											value = match[3]
										case match[4] != "": // Matched unquoted value
											value = match[4]
										}
										params += fmt.Sprintf("`%s`: `%s`,", key, value)
									}
								}
								params += `}`
								*hookContext += `}`
								// fmt.Println(content)
								// fmt.Println("12332112331")
								if ifExpr != "" {
									builder.WriteString(fmt.Sprintf("{{ if %s }} \n", ifExpr))
								}
								builder.WriteString(fmt.Sprintf("{{\n __code__.WriteString(use(`%s`,`%s`, %s, %s)) \n}} \n", main, sub, params, *hookContext))
								if ifExpr != "" {
									builder.WriteString("{{ end }} \n")
								}
							}()
						}
						i = j
						pos = matches[j][1]
						break
					} else if endCmdType != "else" && endCmdType != "redo" {
						count--
					}
				} else if endCmdType != "else" && endCmdType != "redo" {
					count++
				}
			}
		case "redo":
			// 暂时先在这里处理if
			if ifExpr != "" {
				builder.WriteString(fmt.Sprintf("{{ if %s }} \n", ifExpr))
			}
			builder.WriteString(fmt.Sprintf("{{\n __code__.WriteString(redo(`%s`)) \n}} \n", cmdArgs))
			if ifExpr != "" {
				builder.WriteString("{{ end }} \n")
			}
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
	*sql = builder.String()
	// *sql = strings.Join(hookBlocks, "\n") + builder.String()
	return nil
}

func (t *SqlRender) handleCommand(sql *string) {
	// prefix command
	// prefix := `(?:\b(val|each)\b)?`
	// 特殊空格全部转为空格
	spaceRegex := regexp.MustCompile(`[\t\r]`)
	*sql = spaceRegex.ReplaceAllString(*sql, " ")
	command := regexp.MustCompile(`^(.*?)\s*(\bby\b.*?)?(\bif\b.*?)?$`)
	re := regexp.MustCompile(`(?m)^(.*?)--#\s*(.*?)$`)
	eatRe := regexp.MustCompile(`('.*?'|".*?"|[\d\.]+)\s*,?\s*$`)
	*sql = re.ReplaceAllStringFunc(*sql, func(s string) string {
		var pre, middle, post string
		// 指令语句
		matches := re.FindAllStringSubmatch(s, 1)
		// 这里不用判断，一定可以匹配到
		middle = strings.TrimSpace(matches[0][1])
		if middle == "" {
			// 行指令不该由你处理
			if !lineCommandRe.MatchString(s) {
				// 如果不是已知的行指令，说明是函数调用
				return fmt.Sprintf("{{ \n %s \n }} \n", matches[0][2])
			}
			return s
		}
		// 只处理尾指令
		parts := strings.Split(matches[0][2], "$$")
		for _, part := range parts {
			part := strings.TrimSpace(part)
			// 普通指令
			commandSubMatch := command.FindAllStringSubmatch(part, 1)
			// 如果使用了if
			if commandSubMatch[0][3] != "" && strings.HasPrefix(commandSubMatch[0][3], "if") {
				// 使用if
				conditionExpr := commandSubMatch[0][3][2:]
				//fmt.Println(conditionExpr)
				pre += "{{ if " + conditionExpr + " }} \n"
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
						pre += "{{ if " + leftCommand[:len(leftCommand)-1] + " }} \n"
						post += "{{ end }} \n"
						// 去掉？
						leftCommand = leftCommand[:len(leftCommand)-1]
					}
					mainCommand = "{{ val(" + leftCommand + ") }}"
					eatTail = true
				} else if strings.HasPrefix(commandSubMatch[0][1], "each ") {
					leftCommand = leftCommand[5:]
					if strings.HasSuffix(leftCommand, "?") {
						pre += "{{ if " + leftCommand[:len(leftCommand)-1] + " }} \n"
						post += "{{ end }} \n"
						// 去掉？
						leftCommand = leftCommand[:len(leftCommand)-1]
					}
					mainCommand = "{{ each(" + leftCommand + ") }}"
					eatTail = true
				} else {
					if strings.HasSuffix(leftCommand, "?") {
						pre += "{{ if " + leftCommand[:len(leftCommand)-1] + " }} \n"
						post += "{{ end }} \n"
						// 去掉？
						leftCommand = leftCommand[:len(leftCommand)-1]
					}
					mainCommand = "{{ " + leftCommand + " }}"
				}

				if mainCommand == "" {
					mainCommand = "{{ " + commandSubMatch[0][1] + " }}"
				}
			}

			// 使用by的情况
			if commandSubMatch[0][2] != "" {
				if strings.HasPrefix(commandSubMatch[0][2], "by") {
					willBeReplaced := strings.TrimSpace(commandSubMatch[0][2][2:])
					if strings.HasPrefix(willBeReplaced, "/") && strings.HasSuffix(willBeReplaced, "/") {
						replaceRegex := regexp.MustCompile(willBeReplaced[1 : len(willBeReplaced)-1])
						middle = replaceRegex.ReplaceAllStringFunc(middle, func(s string) string {
							subMatch := replaceRegex.FindStringSubmatch(s)
							if len(subMatch) == 0 {
								return s
							}
							result := s
							for i := 1; i < len(subMatch); i++ {
								result = strings.Replace(result, subMatch[i], mainCommand, 1)
							}
							return result
						})
					} else {
						middle = strings.ReplaceAll(middle, willBeReplaced, mainCommand)
					}
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
	if sql, ok := t.sqlMap.blocks[title+":"+subTitle]; ok {
		return sql
	}
	return ""
}

func (t *SqlRender) GetSql(title, subTitle string, data any, handlers ...SqlRenderHandler) (string, []any, error) {
	sql := t.getSql(title, subTitle)
	if sql == "" {
		return "", nil, errors.New("sql not found")
	}
	ctx := &sqlContextItem{
		title:    title,
		subTitle: subTitle,
		params:   make([]any, 0),
		hooks:    []map[string]string{},
		handlers: handlers,
	}
	t.sqlContext.SetContext(ctx)
	defer t.sqlContext.CleanContext()
	inter, err := t.engine.prepareRender(data)
	if err != nil {
		return "", nil, err
	}
	ctx.inters = append(ctx.inters, inter)
	ctx.slotHistories = append(ctx.slotHistories, map[string]string{})
	sql, err = t.engine.doRender(inter, sql)
	if err != nil {
		return "", nil, err
	}
	return sql, ctx.params, ctx.err
}

func (t *SqlRender) GetSqlParams() *[]any {
	ctx := t.sqlContext.GetContext()
	if ctx == nil {
		var arr = make([]any, 0)
		return &arr
	}
	return &ctx.params
}

func NewSqlRender() *SqlRender {
	sqlRender := &SqlRender{
		sqlContext: &sqlContext{},
		sqlMap: &sqlBlocks{
			blocks:    map[string]string{},
			constants: map[int]string{},
		},
	}
	sqlRender.engine = NewTemplateEngine(sqlRender.lib())
	return sqlRender
}
