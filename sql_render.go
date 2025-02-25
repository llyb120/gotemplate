package gotemplate

import (
	"regexp"
	"strings"
	"sync"
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
		sql := strings.TrimSpace(match[2])
		t.sqlMap.blocks[title][subTitle] = sql
	}
	return nil
}

func NewSqlRender() *SqlRender {
	return &SqlRender{
		engine: NewTemplateEngine(),
		sqlMap: &sqlBlocks{
			blocks: map[string]map[string]string{},
		},
	}
}
