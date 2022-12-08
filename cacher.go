package cacher

import (
	"sync"
	"time"
)

const (
	_maxUInt32 = ^uint32(0)
)

// 缓存器
type cacher struct {
	cache     map[any]*cacheItem
	lock      *sync.RWMutex
	maxLength uint
}

// 缓存项
type cacheItem struct {
	value        any
	LastUsedTime time.Time
	UsedTimes    uint32
}

// 创建一个缓存器
// maxLength: 最大缓存长度
func NewCacher(maxLength uint) *cacher {
	if maxLength == 0 {
		maxLength = 100
	}
	
	return &cacher{
		cache:     make(map[any]*cacheItem, maxLength),
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
		item.LastUsedTime = time.Now()
		if item.UsedTimes < _maxUInt32 {
			item.UsedTimes++
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
	full := c.Len() >= c.maxLength

	if !full {
		c.lock.Lock()
		c.cache[key] = &cacheItem{
			value:        value,
			LastUsedTime: time.Now(),
			UsedTimes:    1,
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
			lastTime = item.LastUsedTime
			usedTimes = item.UsedTimes
			lastKey = key
			break
		}
		for key, item := range c.cache {
			if item.LastUsedTime.Before(lastTime) && item.UsedTimes <= usedTimes {
				lastTime = item.LastUsedTime
				usedTimes = item.UsedTimes
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
		item.LastUsedTime = time.Now()
		if item.UsedTimes < _maxUInt32 {
			item.UsedTimes++
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
	c.cache = make(map[any]*cacheItem, c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *cacher) Len() uint {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return uint(len(c.cache))
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
	defer c.lock.RUnlock()
	return m
}
