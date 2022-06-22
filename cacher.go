package cacher

import (
	"sync"
	"time"
)

var (
	emptyCache = make(map[interface{}]CacheItem, 0)
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
		cache:        emptyCache,
		maxLength:    maxLength,
	}
}

// 添加/更新一个缓存项
func (c *Cacher) Update(key, value interface{}) {
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
	c.lock.RLock()
	full := len(c.cache) >= c.maxLength
	c.lock.RUnlock()

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

		return item.value, true
	}
	return nil, false
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
	c.cache = emptyCache
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *Cacher) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.cache)
}

// 获取所有缓存项
func (c *Cacher) All() []interface{} {
	items := []interface{}{}
	for _, item := range c.cache {
		items = append(items, item.value)
	}
	return items
}