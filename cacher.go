package golrucacher

import (
	"sync"
)

type cacheMap map[any]*CacheItem

// 缓存项
type CacheItem struct {
	Value    any
	LastTime int64
	Times    int
}

// 缓存器
type cacher struct {
	caches    cacheMap
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
		caches:    make(cacheMap, maxLength),
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
		c.caches[key].Value = value
		defer c._update(key)
		return
	}

	// key不存在，添加。
START:
	if c.IsFull() {
		k, _, _ := c.Activest()
		c.lock.Lock()
		delete(c.caches, k)
		c.lock.Unlock()

		goto START
	}

	c.lock.Lock()
	c.caches[key] = &CacheItem{
		Value:    value,
		LastTime: _unixNano(),
		Times:    1,
	}
	c.lock.Unlock()
}

// 返回最早添加、最少使用的项目的键、添加时间、使用次数
func (c *cacher) Activest() (lastKey any, lastTime int64, times int) {
	c.lock.RLock()
	for key, item := range c.caches {
		if lastKey == nil || item.LastTime <= lastTime && item.Times <= times {
			lastTime = item.LastTime
			times = item.Times
			lastKey = key
		}
	}
	c.lock.RUnlock()
	return
}

// 返回最晚添加、最多使用的项目的键、添加时间、使用次数
func (c *cacher) Laziest() (lastKey any, lastTime int64, times int) {
	c.lock.RLock()
	for key, item := range c.caches {
		if lastKey == nil || item.LastTime >= lastTime && item.Times >= times {
			lastTime = item.LastTime
			times = item.Times
			lastKey = key
		}
	}
	c.lock.RUnlock()
	return
}

// 获取一个缓存项
func (c *cacher) Get(key any) (any, bool) {
	c.lock.RLock()
	item, ok := c.caches[key]
	c.lock.RUnlock()

	if ok {
		defer c._update(key)
		return item.Value, ok
	}
	return nil, false
}

func (c *cacher) _update(key any) {
	c.lock.Lock()
	c.caches[key].Times++
	c.caches[key].LastTime = _unixNano()
	c.lock.Unlock()
}

// 删除缓存项
func (c *cacher) Delete(key any) {
	c.lock.Lock()
	delete(c.caches, key)
	c.lock.Unlock()
}

func (c *cacher) DeleteAll(keys []any) {
	c.lock.Lock()
	for _, key := range keys {
		delete(c.caches, key)
	}
	c.lock.Unlock()
}

func (c *cacher) DeleteAllFn(fn func(key any, item *CacheItem) bool) {
	c.lock.Lock()
	for key, item := range c.caches {
		if fn(key, item) {
			delete(c.caches, key)
		}
	}
	c.lock.Unlock()
}

// 清空
func (c *cacher) Clear() {
	c.lock.Lock()
	c.caches = make(cacheMap, c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *cacher) Len() int {
	return len(c.caches)
}

// 是否存满
func (c *cacher) IsFull() bool {
	return c.Len() >= c.maxLength
}

// 获取所有值
func (c *cacher) Values() []any {
	c.lock.RLock()
	items := make([]any, 0, len(c.caches))
	for _, item := range c.caches {
		items = append(items, item.Value)
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
		m[key] = item.Value
	}
	c.lock.RUnlock()
	return m
}
