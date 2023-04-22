package golrucacher

import (
	"sync"
)

type cacheMap[T any] map[any]*CacheItem[T]
type kvMap[T any] map[any]*T

// 缓存项
type CacheItem[T any] struct {
	Value    *T
	LastTime float64
	Times    int
}

// 缓存器
type cacher[T any] struct {
	caches    cacheMap[T]
	lock      *sync.RWMutex
	maxLength int
}

// 创建一个缓存器
// maxLength: 最大缓存长度
func NewCacher[T any](maxLength int) *cacher[T] {
	if maxLength <= 0 {
		panic("maxLength must be greater than 0")
	}

	return &cacher[T]{
		caches:    make(cacheMap[T], maxLength),
		maxLength: maxLength,
		lock:      new(sync.RWMutex),
	}
}

// 添加/更新一个缓存项
func (c *cacher[T]) Set(key any, value *T) {
	// key存在，更新
	c.lock.RLock()
	_, have := c.caches[key]
	c.lock.RUnlock()
	if have {
		c.caches[key].Value = value
		defer c.update(key)
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
	c.caches[key] = &CacheItem[T]{
		Value:    value,
		LastTime: getTime(),
		Times:    1,
	}
	c.lock.Unlock()
}

// 返回最早添加、最少使用的项目的键、添加时间、使用次数
func (c *cacher[T]) Activest() (lastKey any, lastTime float64, times int) {
	c.lock.RLock()
	for key, item := range c.caches {
		if lastKey == nil || item.LastTime <= lastTime || item.Times <= times {
			lastTime = item.LastTime
			times = item.Times
			lastKey = key
		}
	}
	c.lock.RUnlock()
	return
}

// 返回最晚添加、最多使用的项目的键、添加时间、使用次数
func (c *cacher[T]) Laziest() (lastKey any, lastTime float64, times int) {
	c.lock.RLock()
	for key, item := range c.caches {
		if lastKey == nil || item.LastTime >= lastTime || item.Times >= times {
			lastTime = item.LastTime
			times = item.Times
			lastKey = key
		}
	}
	c.lock.RUnlock()
	return
}

// 获取一个缓存项
func (c *cacher[T]) Get(key any) (*T, bool) {
	c.lock.RLock()
	item, ok := c.caches[key]
	c.lock.RUnlock()

	if ok {
		defer c.update(key)
		return item.Value, true
	}
	return nil, false
}

func (c *cacher[T]) update(key any) {
	_, ok := c.caches[key]
	if !ok {
		return
	}
	c.lock.Lock()
	c.caches[key].Times++
	c.caches[key].LastTime = getTime()
	c.lock.Unlock()
}

// 删除缓存项
func (c *cacher[T]) Delete(key any) {
	c.lock.Lock()
	delete(c.caches, key)
	c.lock.Unlock()
}

func (c *cacher[T]) DeleteAll(keys []any) {
	c.lock.Lock()
	for _, key := range keys {
		delete(c.caches, key)
	}
	c.lock.Unlock()
}

func (c *cacher[T]) DeleteAllFn(fn func(key any, item *CacheItem[T]) bool) {
	c.lock.Lock()
	for key, item := range c.caches {
		if fn(key, item) {
			delete(c.caches, key)
		}
	}
	c.lock.Unlock()
}

// 清空
func (c *cacher[T]) Clear() {
	c.lock.Lock()
	c.caches = make(cacheMap[T], c.maxLength)
	c.lock.Unlock()
}

// 获取缓存项数量
func (c *cacher[T]) Len() int {
	return len(c.caches)
}

// 是否存满
func (c *cacher[T]) IsFull() bool {
	return c.Len() >= c.maxLength
}

// 获取所有值
func (c *cacher[T]) Values() []*T {
	c.lock.RLock()
	items := make([]*T, 0, len(c.caches))
	for _, item := range c.caches {
		items = append(items, item.Value)
	}
	c.lock.RUnlock()
	return items
}

// 获取所有键
func (c *cacher[T]) Keys() []any {
	c.lock.RLock()
	keys := make([]any, 0, len(c.caches))
	for key := range c.caches {
		keys = append(keys, key)
	}
	c.lock.RUnlock()
	return keys
}

func (c *cacher[T]) Map() kvMap[T] {
	c.lock.RLock()
	m := make(kvMap[T], len(c.caches))
	for key, item := range c.caches {
		m[key] = item.Value
	}
	c.lock.RUnlock()
	return m
}

func (c *cacher[T]) changeLen(len int) (overflow cacheMap[T]) {
	if len >= c.maxLength {
		c.maxLength = len
		return nil
	}

	overflowCount := c.maxLength - len
	overflow = make(cacheMap[T], overflowCount)
	for i := 0; i < overflowCount; i++ {
		k, _, _ := c.Activest()
		overflow[k] = c.caches[k]
		c.Delete(k)
	}
	c.maxLength = len
	return
}

func (c *cacher[T]) addCacheMap(m cacheMap[T]) {
	c.lock.Lock()
	for k, v := range m {
		c.caches[k] = v
	}
	c.lock.Unlock()
}
