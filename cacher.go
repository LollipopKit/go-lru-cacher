package cacher

import (
	"sync"
	"time"
)

// 缓存器
type Cacher struct {
	cache        map[interface{}]CacheItem
	lock         sync.RWMutex
	maxLength    int
}

// 缓存项
type CacheItem struct {
	value        interface{}
	LastUsedTime time.Time
	UsedTimes    int
}

// 创建一个缓存器
// maxLength: 最大缓存长度
func NewCacher(maxLength int) *Cacher {
	if maxLength <= 0 {
		maxLength = 100
	}

	return &Cacher{
		cache:        make(map[interface{}]CacheItem, maxLength),
		maxLength:    maxLength,
	}
}

// 添加/更新一个缓存项
func (c *Cacher) Set(key, value interface{}) {
	// key存在，更新
	c.lock.RLock()
	item, have := c.cache[key]
	c.lock.RUnlock()
	if have {
		item.value = value
		item.LastUsedTime = time.Now()
		item.UsedTimes++

		c.lock.Lock()
		c.cache[key] = item
		c.lock.Unlock()
		return
	}

	// key不存在，添加
START:
	full := c.Len() >= c.maxLength

	if !full {
		c.lock.Lock()
		c.cache[key] = CacheItem{
			value:        value,
			LastUsedTime: time.Now(),
			UsedTimes:    1,
		}
		c.lock.Unlock()
	} else {
		var lastTime time.Time
		var usedTimes int
		var lastKey interface{}

		c.lock.RLock()
		for key, value := range c.cache {
			lastTime = value.LastUsedTime
			usedTimes = value.UsedTimes
			lastKey = key
			break
		}
		for idx, item := range c.cache {
			if item.LastUsedTime.Before(lastTime) && item.UsedTimes <= usedTimes {
				lastTime = item.LastUsedTime
				usedTimes = item.UsedTimes
				lastKey = idx
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
func (c *Cacher) Get(key interface{}) (interface{}, bool) {
	c.lock.RLock()
	item, ok := c.cache[key]
	c.lock.RUnlock()

	if ok {
		item.LastUsedTime = time.Now()
		item.UsedTimes++

		c.lock.Lock()
		c.cache[key] = item
		c.lock.Unlock()
	}
	return item.value, ok
}

// 删除缓存项
func (c *Cacher) Delete(key interface{}) {
	c.lock.Lock()
	delete(c.cache, key)
	c.lock.Unlock()
}

// 清空
func (c *Cacher) Clear() {
	c.lock.Lock()
	c.cache = make(map[interface{}]CacheItem, c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *Cacher) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.cache)
}

// 获取所有值
func (c *Cacher) Values() []interface{} {
	items := []interface{}{}
	c.lock.RLock()
	for _, item := range c.cache {
		items = append(items, item.value)
	}
	c.lock.RUnlock()
	return items
}

// 获取所有键
func (c *Cacher) Keys() []interface{} {
	keys := []interface{}{}
	c.lock.RLock()
	for key := range c.cache {
		keys = append(keys, key)
	}
	c.lock.RUnlock()
	return keys
}