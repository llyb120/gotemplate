package gotemplate

import (
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

// 对反引号进行转义
func escapeBacktick(content *string) {
	// 由于go的标签内无法直接包含反引号，所以只能用字符串拼接的方式进行转义
	*content = strings.ReplaceAll(*content, "`", "`+\"`\"+`")
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
