package gotemplate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	sql, params, err := sqlRender.GetSql("test", "sql1", map[string]any{
		"x":   "foo",
		"arr": []any{"1", "2"},
		"a":   4,
		"mp": map[string]any{
			"k": 1,
			"m": 2,
		},
		"b": true,
	})

	fmt.Println(sql, params, err)
}
