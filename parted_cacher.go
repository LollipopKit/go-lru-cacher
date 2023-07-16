package golrucacher

type PartedCacher[T any] struct {
	// 如果有部分数据一直被读取，可以使用 `lazy / active` 分区的缓存器，这样可以提高缓存命中率。
	// ### 逻辑
	// - 先把 `active` 写满，再写 `lazy`
	// - `active` 写满时写入，先将` active` 的最近、使用最多的缓存项“移动”到 `lazy`，再写入 `active`
	// - `active` 写满时写入，`lazy` 写满，删除 `lazy` 中最早、使用最少的缓存项，移动 `active` 的最近、使用最多的缓存项到 `lazy`，再写入 `active`

	Active     *Cacher[T] // “删”时优先处理该部分
	Lazy       *Cacher[T] // “查”时优先处理该部分
	maxLen     int
	activeRate float64 // 其值为 activeLen / maxLength
}

// [activeRate]: 0 < activeRate < 1, activeRate = activeLen / maxLength
func NewPartedCacher[T any](maxLength int, activeRate float64) *PartedCacher[T] {
	if activeRate <= 0 || activeRate >= 1 {
		panic("activeRate must be in (0, 1)")
	}
	a, l := calcMaxLength(maxLength, activeRate)

	return &PartedCacher[T]{
		Active:     NewCacher[T](a),
		Lazy:       NewCacher[T](l),
		maxLen:     maxLength,
		activeRate: activeRate,
	}
}

func (c *PartedCacher[T]) AdjustRate(rate float64) bool {
	diffRate := c.activeRate - rate
	rate = c.activeRate - diffRate/2

	aLen, lLen := calcMaxLength(c.maxLen, rate)
	if aLen != c.Active.maxLength {
		if aLen > c.Active.maxLength {
			c.Active.changeLen(aLen)
			c.Active.addCacheMap(c.Lazy.changeLen(lLen))
		} else {
			c.Lazy.changeLen(lLen)
			c.Lazy.addCacheMap(c.Active.changeLen(aLen))
		}
		c.activeRate = rate
		return true
	}
	return false
}

func (c *PartedCacher[T]) moveActive2Lazy(keyA, keyL any) bool {
	cacheItemA, ok := c.Active.caches[keyA]
	if !ok {
		return false
	}
	cacheItemL, ok := c.Lazy.caches[keyL]
	if !ok {
		return false
	}

	c.Lazy.Delete(keyL)
	c.Active.Delete(keyA)

	c.Lazy.lock.Lock()
	c.Lazy.caches[keyA] = cacheItemA
	c.Lazy.lock.Unlock()

	c.Active.lock.Lock()
	c.Active.caches[keyL] = cacheItemL
	c.Active.lock.Unlock()

	return true
}

func (c *PartedCacher[T]) moveLazy2Active(keyA, keyL any) bool {
	cacheItemA, ok := c.Active.caches[keyA]
	if !ok {
		return false
	}
	cacheItemL, ok := c.Lazy.caches[keyL]
	if !ok {
		return false
	}

	c.Lazy.Delete(keyL)
	c.Active.Delete(keyA)

	c.Active.lock.Lock()
	c.Active.caches[keyL] = cacheItemL
	c.Active.lock.Unlock()

	c.Lazy.lock.Lock()
	c.Lazy.caches[keyA] = cacheItemA
	c.Lazy.lock.Unlock()

	return true
}

func (c *PartedCacher[T]) Set(key any, value *T) {
	// 不能使用 `Cacher.Get(key)`
	// 因为会导致其 `LastTime` 和 `Times` 更新
	if c.Lazy.caches[key] != nil {
		c.Lazy.Set(key, value)
		return
	}

	// `active` 未满
	if !c.Active.IsFull() {
		c.Active.Set(key, value)
		return
	}

	delKeyInActive, aTime, aTimes := c.Active.Laziest()
	v, ok := c.Active.Get(delKeyInActive)

	// active满，但lazy未满：
	if !c.Lazy.IsFull() {
		if ok {
			// 1、移动active的最近、使用最多的缓存项到lazy
			c.Lazy.Set(delKeyInActive, v)
			// 2、删除active中最早添加、使用最多的缓存项
			c.Active.Delete(delKeyInActive)
		}
		// 3、将新增项目添加到active
		c.Active.Set(key, value)
		return
	}

	delKeyInLazy, lTime, lTimes := c.Lazy.Activest()

	// lazy满
	// 且lazy中最早、使用最少的缓存项
	// 比
	// active中最近、使用最多的缓存项更早、使用更少
	if lTime <= aTime || lTimes <= aTimes { // FIFO：先进先出，所以包含等于
		c.moveActive2Lazy(delKeyInActive, delKeyInLazy)
		// 将新增项目添加到active
		c.Active.Set(key, value)
		return
	}

	// lazy满
	// 但lazy中最早、使用最少的缓存项
	// 比
	// active中最近、使用最多的缓存项更近、使用更多
	// 所以，只需将新增项目添加到active
	c.Active.Set(key, value)
}

func (c *PartedCacher[T]) Get(key any) (*T, bool) {
	v, ok := c.Lazy.Get(key)
	if ok {
		return v, ok
	}

	value, ok := c.Active.Get(key)
	if ok {
		delKeyInActive, lTime, lTimes := c.Active.Laziest()
		delKeyInLazy, aTime, aTimes := c.Lazy.Activest()
		if aTime <= lTime || aTimes <= lTimes { // FIFO：先进先出，所以包含等于
			// 1、删除lazy中最早、使用最少的缓存项
			c.Lazy.Delete(delKeyInLazy)
			// 2、移动active的最近、使用最多的缓存项到lazy
			c.Lazy.Set(delKeyInActive, value)
			c.Active.Delete(delKeyInActive)
		}
	}
	return value, ok
}

func (c *PartedCacher[T]) Delete(key any) {
	c.Active.Delete(key)
	c.Lazy.Delete(key)
}

func (c *PartedCacher[T]) DeleteAll(keys []any) {
	c.Active.DeleteAll(keys)
	c.Lazy.DeleteAll(keys)
}

func (c *PartedCacher[T]) DeleteAllFn(fn func(key any, value *CacheItem[T]) bool) {
	c.Active.DeleteAllFn(fn)
	c.Lazy.DeleteAllFn(fn)
}

func (c *PartedCacher[T]) IsFull() bool {
	return c.Active.IsFull() && c.Lazy.IsFull()
}

func (c *PartedCacher[T]) Clear() {
	c.Active.Clear()
	c.Lazy.Clear()
}

func (c *PartedCacher[T]) Len() int {
	return c.Active.Len() + c.Lazy.Len()
}

func (c *PartedCacher[T]) Keys() []any {
	return append(c.Active.Keys(), c.Lazy.Keys()...)
}

func (c *PartedCacher[T]) Values() []*T {
	return append(c.Active.Values(), c.Lazy.Values()...)
}

func (c *PartedCacher[T]) Map() kvMap[T] {
	aMap := c.Active.Map()
	lMap := c.Lazy.Map()
	for k := range lMap {
		aMap[k] = lMap[k]
	}
	return aMap
}

func (c *PartedCacher[T]) PartedMap() map[string]kvMap[T] {
	aMap := c.Active.Map()
	lMap := c.Lazy.Map()
	m := map[string]kvMap[T]{"active": {}, "lazy": {}}
	for k := range lMap {
		m["lazy"][k] = lMap[k]
	}
	for k := range aMap {
		m["active"][k] = aMap[k]
	}
	return m
}
