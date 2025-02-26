package gotemplate

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

// 对反引号进行转义
func escapeBacktick(content string) string {
	// 由于go的标签内无法直接包含反引号，所以只能用字符串拼接的方式进行转义
	return strings.ReplaceAll(content, "`", "`+\"`\"+`")
}

// 对代码进行转义
func encodeCode(content *string) {
	re := regexp.MustCompile(`\{\{|\}\}`)
	*content = re.ReplaceAllStringFunc(*content, func(s string) string {
		if s == "{{" {
			return "@{"
		} else if s == "}}" {
			return "@}"
		}
		return s
	})
}

func decodeCode(content *string) {
	re := regexp.MustCompile(`@\{|@\}`)
	*content = re.ReplaceAllStringFunc(*content, func(s string) string {
		if s == "@{" {
			return "{{"
		} else if s == "@}" {
			return "}}"
		}
		return s
	})
}

// 错误组
type ErrGroup struct {
	errgroup.Group
}

func (e *ErrGroup) Go(fn func() error) {
	e.Group.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		return fn()
	})
}
