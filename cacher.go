package cacher

import (
	"sync"
	"time"
)

const (
	_maxUInt32 = ^uint32(0)
)

type caches map[any]*cacheItem

// 缓存项
type cacheItem struct {
	value    any
	lastTime time.Time
	times    uint32
}

// 缓存器
type cacher struct {
	cache     caches
	lock      *sync.RWMutex
	maxLength uint32
}


// 创建一个缓存器
// maxLength: 最大缓存长度
func NewCacher(maxLength uint32) *cacher {
	if maxLength == 0 {
		panic("maxLength must be greater than 0")
	}

	return &cacher{
		cache:     make(caches, maxLength),
		maxLength: maxLength,
		lock:      new(sync.RWMutex),
	}
}

// 添加/更新一个缓存项
func (c *cacher) Set(key, value any) {
	// key存在，更新
	c.lock.RLock()
	item, have := c.cache[key]
	c.lock.RUnlock()
	if have {
		item.value = value
		item.lastTime = time.Now()
		if item.times < _maxUInt32 {
			item.times++
		}

		c.lock.Lock()
		c.cache[key] = item
		c.lock.Unlock()
		return
	}

	// key不存在，添加。
	// 因为每次插入时都会检查是否已满，所以每次插入时，最多需要清除一个缓存项。
	// 不需要一个deleteList，在一次for内删除掉所有过期的缓存项。
START:
	full := uint32(c.Len()) >= c.maxLength

	if !full {
		c.lock.Lock()
		c.cache[key] = &cacheItem{
			value:    value,
			lastTime: time.Now(),
			times:    1,
		}
		c.lock.Unlock()
	} else {
		var lastTime time.Time
		var usedTimes uint32
		var lastKey any

		c.lock.RLock()
		// 先使用一次for循环，而不是只有一个for
		// 这样性能反而更好
		for key, item := range c.cache {
			lastTime = item.lastTime
			usedTimes = item.times
			lastKey = key
			break
		}
		for key, item := range c.cache {
			if item.lastTime.Before(lastTime) && item.times <= usedTimes {
				lastTime = item.lastTime
				usedTimes = item.times
				lastKey = key
			}
		}
		c.lock.RUnlock()

		c.lock.Lock()
		delete(c.cache, lastKey)
		c.lock.Unlock()

		goto START
	}
}

// 获取一个缓存项
func (c *cacher) Get(key any) (any, bool) {
	c.lock.RLock()
	item, ok := c.cache[key]
	c.lock.RUnlock()

	if ok {
		item.lastTime = time.Now()
		if item.times < _maxUInt32 {
			item.times++
		}

		c.lock.Lock()
		c.cache[key] = item
		c.lock.Unlock()

		return item.value, ok
	}
	return nil, false
}

// 删除缓存项
func (c *cacher) Delete(key any) {
	c.lock.Lock()
	delete(c.cache, key)
	c.lock.Unlock()
}

// 清空
func (c *cacher) Clear() {
	c.lock.Lock()
	c.cache = make(caches, c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *cacher) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.cache)
}

// 获取所有值
func (c *cacher) Values() []any {
	c.lock.RLock()
	items := make([]any, 0, len(c.cache))
	for _, item := range c.cache {
		items = append(items, item.value)
	}
	c.lock.RUnlock()
	return items
}

// 获取所有键
func (c *cacher) Keys() []any {
	c.lock.RLock()
	keys := make([]any, 0, len(c.cache))
	for key := range c.cache {
		keys = append(keys, key)
	}
	c.lock.RUnlock()
	return keys
}

func (c *cacher) Map() map[any]any {
	c.lock.RLock()
	m := make(map[any]any, len(c.cache))
	for key, item := range c.cache {
		m[key] = item.value
	}
	c.lock.RUnlock()
	return m
}
