package gotemplate

import "strings"

// 对反引号进行转义
func escapeBacktick(content string) string {
	// 由于go的标签内无法直接包含反引号，所以只能用字符串拼接的方式进行转义
	return strings.ReplaceAll(content, "`", "`+\"`\"+`")
}
