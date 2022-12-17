package cacher

import (
	"sync"
)

const (
	_maxInt = int(^uint(0) >> 1)
)

type _cacheMap map[any]*cacheItem

// 缓存项
type cacheItem struct {
	value    any
	lastTime int64
	times    int
}

// 缓存器
type cacher struct {
	caches    _cacheMap
	lock      *sync.RWMutex
	maxLength int
}

// 创建一个缓存器
// maxLength: 最大缓存长度
func NewCacher(maxLength int) *cacher {
	if maxLength <= 0 {
		panic("maxLength must be greater than 0")
	}

	return &cacher{
		caches:    make(_cacheMap, maxLength),
		maxLength: maxLength,
		lock:      new(sync.RWMutex),
	}
}

// 添加/更新一个缓存项
func (c *cacher) Set(key, value any) {
	// key存在，更新
	c.lock.RLock()
	_, have := c.caches[key]
	c.lock.RUnlock()
	if have {
		c.caches[key].value = value
		defer c._update(key)
		return
	}

	// key不存在，添加。
START:
	full := c.Len() >= c.maxLength

	if full {
		var lastTime int64
		var usedTimes int
		var lastKey any

		c.lock.RLock()
		for key, item := range c.caches {
			if lastKey == nil || item.lastTime <= lastTime && item.times <= usedTimes {
				lastTime = item.lastTime
				usedTimes = item.times
				lastKey = key
			}
		}
		c.lock.RUnlock()

		c.lock.Lock()
		delete(c.caches, lastKey)
		c.lock.Unlock()

		goto START
	}

	c.lock.Lock()
	c.caches[key] = &cacheItem{
		value:    value,
		lastTime: _unixNano(),
		times:    1,
	}
	c.lock.Unlock()
}

// 获取一个缓存项
func (c *cacher) Get(key any) (any, bool) {
	c.lock.RLock()
	item, ok := c.caches[key]
	c.lock.RUnlock()

	if ok {
		defer c._update(key)
		return item.value, ok
	}
	return nil, false
}

func (c *cacher) _update(key any) {
	c.lock.Lock()
	c.caches[key].times++
	c.caches[key].lastTime = _unixNano()
	c.lock.Unlock()
}

// 删除缓存项
func (c *cacher) Delete(key any) {
	c.lock.Lock()
	delete(c.caches, key)
	c.lock.Unlock()
}

// 清空
func (c *cacher) Clear() {
	c.lock.Lock()
	c.caches = make(_cacheMap, c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *cacher) Len() int {
	return len(c.caches)
}

// 获取所有值
func (c *cacher) Values() []any {
	c.lock.RLock()
	items := make([]any, 0, len(c.caches))
	for _, item := range c.caches {
		items = append(items, item.value)
	}
	c.lock.RUnlock()
	return items
}

// 获取所有键
func (c *cacher) Keys() []any {
	c.lock.RLock()
	keys := make([]any, 0, len(c.caches))
	for key := range c.caches {
		keys = append(keys, key)
	}
	c.lock.RUnlock()
	return keys
}

func (c *cacher) Map() map[any]any {
	c.lock.RLock()
	m := make(map[any]any, len(c.caches))
	for key, item := range c.caches {
		m[key] = item.value
	}
	c.lock.RUnlock()
	return m
}
