package gotemplate

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestSqlRender_Call(t *testing.T) {
	refVal := reflect.ValueOf(_str)
	result := refVal.Call([]reflect.Value{
		reflect.Zero(reflect.TypeOf((*int)(nil)).Elem()),
	})
	fmt.Println(result)
}

func TestSqlRender_Scan(t *testing.T) {
	var g ErrGroup
	dir := "./examples"
	sqlRender := NewSqlRender()
	if err := sqlRender.Scan(func(handler ScanHandler) error {
		return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}
			g.Go(func() error {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				return handler(path, string(content))
			})
			return nil
		})
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}

	sql, params, err := sqlRender.GetSql("test", "test3", map[string]any{
		"x":   "foo",
		"arr": []any{"1", "2"},
		"a":   4,
		"mp":  map[string]any{},
		"b":   true,
		"test": func() bool {
			return true
		},
		"Items":  []any{"1", "2", "3"},
		"Items2": []any{"1", "2", "3", "4", "5"},
		"Test": func() string {
			println("Test")
			return "123"
		},
	}, func(phase SqlRenderPhase, context SqlHandlerContext, sql *string, args *[]any) error {
		if phase == ON_SLOT_RENDER {
			*sql += "foo bar"
			// 	*sql = strings.ReplaceAll(*sql, "{{x}}", "{{.x}}")
		}
		if phase == ON_REDO_RENDER {
			*sql += "shou me the monty"
		}
		return nil
	})

	fmt.Println(sql, params, err)
}

func TestUseRegex(t *testing.T) {
	re := regexp.MustCompile(`^(?:(.*?)\.)?(.*?)\s*(?:\s{1,}as\s{1,}(.*?))?\s*$`)
	arr := re.FindAllStringSubmatch("pc_console_games_topcharts_base as period_data", 1)
	t.Log(arr)
}
