package gotemplate

import "sync"

type parsedCache struct {
	sync.RWMutex
	cache map[string]string
}

func (c *parsedCache) GetIfNotExist(key string, fn func() string) string {
	res := (func() string {
		c.RLock()
		defer c.RUnlock()
		if value, ok := c.cache[key]; ok {
			return value
		}
		return ""
	})()
	if res != "" {
		return res
	}
	c.Lock()
	defer c.Unlock()
	// double check
	if value, ok := c.cache[key]; ok {
		return value
	}
	res = fn()
	c.cache[key] = res
	return res
}
