package gotemplate

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/sync/errgroup"
)

// 对反引号进行转义
func escapeBacktick(content *string) {
	// 由于go的标签内无法直接包含反引号，所以只能用字符串拼接的方式进行转义
	*content = strings.ReplaceAll(*content, "`", "`+\"`\"+`")
}

// 对代码进行转义
func encodeCode(content *string) {
	// 直接base64
	*content = base64.StdEncoding.EncodeToString([]byte(*content))
	// re := regexp.MustCompile(`\{\{|\}\}`)
	// *content = re.ReplaceAllStringFunc(*content, func(s string) string {
	// 	if s == "{{" {
	// 		return "@{"
	// 	} else if s == "}}" {
	// 		return "@}"
	// 	}
	// 	return s
	// })
}

func decodeCode(content *string) {
	// 直接base64解码
	decoded, err := base64.StdEncoding.DecodeString(*content)
	if err != nil {
		fmt.Println("decodeCode error", err)
		return
	}
	*content = string(decoded)
	// re := regexp.MustCompile(`@\{|@\}`)
	// *content = re.ReplaceAllStringFunc(*content, func(s string) string {
	// 	if s == "@{" {
	// 		return "{{"
	// 	} else if s == "@}" {
	// 		return "}}"
	// 	}
	// 	return s
	// })
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
