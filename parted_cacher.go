package golrucacher

type PartedCacher struct {
	// 如果有部分数据一直被读取，可以使用 `lazy / active` 分区的缓存器，这样可以提高缓存命中率。
	// ### 逻辑
	// - 先把 `active` 写满，再写 `lazy`
	// - `active` 写满时写入，先将` active` 的最近、使用最多的缓存项“移动”到 `lazy`，再写入 `active`
	// - `active` 写满时写入，`lazy` 写满，删除 `lazy` 中最早、使用最少的缓存项，移动 `active` 的最近、使用最多的缓存项到 `lazy`，再写入 `active`

	ActiveCacher *cacher // “删”时优先处理该部分
	LazyCacher   *cacher // “查”时优先处理该部分
	maxLen       int
	activeRate   float64 // 其值为 activeLen / maxLength
}

// [activeRate]: 0 < activeRate < 1, activeRate = activeLen / maxLength
func NewPartedCacher(maxLength int, activeRate float64) *PartedCacher {
	if activeRate <= 0 || activeRate >= 1 {
		panic("activeRate must be in (0, 1)")
	}
	a, l := calcMaxLength(maxLength, activeRate)

	return &PartedCacher{
		ActiveCacher: NewCacher(a),
		LazyCacher:   NewCacher(l),
		maxLen:       maxLength,
		activeRate:   activeRate,
	}
}

func (c *PartedCacher) AdjustRate(rate float64) bool {
	diffRate := c.activeRate - rate
	rate = c.activeRate - diffRate/2

	aLen, lLen := calcMaxLength(c.maxLen, rate)
	if aLen != c.ActiveCacher.maxLength {
		if aLen > c.ActiveCacher.maxLength {
			c.ActiveCacher.changeLen(aLen)
			c.ActiveCacher.addCacheMap(c.LazyCacher.changeLen(lLen))
		} else {
			c.LazyCacher.changeLen(lLen)
			c.LazyCacher.addCacheMap(c.ActiveCacher.changeLen(aLen))
		}
		println()
		c.activeRate = rate
		return true
	}
	return false
}

func (c *PartedCacher) moveActive2Lazy(keyA, keyL any) bool {
	cacheItemA, ok := c.ActiveCacher.caches[keyA]
	if !ok {
		return false
	}
	cacheItemL, ok := c.LazyCacher.caches[keyL]
	if !ok {
		return false
	}

	c.LazyCacher.Delete(keyL)
	c.ActiveCacher.Delete(keyA)

	c.LazyCacher.lock.Lock()
	c.LazyCacher.caches[keyA] = cacheItemA
	c.LazyCacher.lock.Unlock()

	c.ActiveCacher.lock.Lock()
	c.ActiveCacher.caches[keyL] = cacheItemL
	c.ActiveCacher.lock.Unlock()

	return true
}

func (c *PartedCacher) moveLazy2Active(keyA, keyL any) bool {
	cacheItemA, ok := c.ActiveCacher.caches[keyA]
	if !ok {
		return false
	}
	cacheItemL, ok := c.LazyCacher.caches[keyL]
	if !ok {
		return false
	}

	c.LazyCacher.Delete(keyL)
	c.ActiveCacher.Delete(keyA)

	c.ActiveCacher.lock.Lock()
	c.ActiveCacher.caches[keyL] = cacheItemL
	c.ActiveCacher.lock.Unlock()

	c.LazyCacher.lock.Lock()
	c.LazyCacher.caches[keyA] = cacheItemA
	c.LazyCacher.lock.Unlock()

	return true
}

func (c *PartedCacher) Set(key, value any) {
	// 不能使用 `Cacher.Get(key)`
	// 因为会导致其 `LastTime` 和 `Times` 更新
	if c.LazyCacher.caches[key] != nil {
		c.LazyCacher.Set(key, value)
		return
	}

	// `active` 未满
	if !c.ActiveCacher.IsFull() {
		c.ActiveCacher.Set(key, value)
		return
	}

	delKeyInActive, aTime, aTimes := c.ActiveCacher.Laziest()
	v, ok := c.ActiveCacher.Get(delKeyInActive)

	// active满，但lazy未满：
	if !c.LazyCacher.IsFull() {
		if ok {
			// 1、移动active的最近、使用最多的缓存项到lazy
			c.LazyCacher.Set(delKeyInActive, v)
			// 2、删除active中最早添加、使用最多的缓存项
			c.ActiveCacher.Delete(delKeyInActive)
		}
		// 3、将新增项目添加到active
		c.ActiveCacher.Set(key, value)
		return
	}

	delKeyInLazy, lTime, lTimes := c.LazyCacher.Activest()

	// lazy满
	// 且lazy中最早、使用最少的缓存项
	// 比
	// active中最近、使用最多的缓存项更早、使用更少
	if lTime <= aTime || lTimes <= aTimes { // FIFO：先进先出，所以包含等于
		c.moveActive2Lazy(delKeyInActive, delKeyInLazy)
		// 将新增项目添加到active
		c.ActiveCacher.Set(key, value)
		return
	}

	// lazy满
	// 但lazy中最早、使用最少的缓存项
	// 比
	// active中最近、使用最多的缓存项更近、使用更多
	// 所以，只需将新增项目添加到active
	c.ActiveCacher.Set(key, value)
}

func (c *PartedCacher) Get(key any) (any, bool) {
	v, ok := c.LazyCacher.Get(key)
	if ok {
		return v, ok
	}

	value, ok := c.ActiveCacher.Get(key)
	if ok {
		delKeyInActive, lTime, lTimes := c.ActiveCacher.Laziest()
		delKeyInLazy, aTime, aTimes := c.LazyCacher.Activest()
		if aTime <= lTime || aTimes <= lTimes { // FIFO：先进先出，所以包含等于
			// 1、删除lazy中最早、使用最少的缓存项
			c.LazyCacher.Delete(delKeyInLazy)
			// 2、移动active的最近、使用最多的缓存项到lazy
			c.LazyCacher.Set(delKeyInActive, value)
			c.ActiveCacher.Delete(delKeyInActive)
		}
	}
	return value, ok
}

func (c *PartedCacher) Delete(key any) {
	c.ActiveCacher.Delete(key)
	c.LazyCacher.Delete(key)
}

func (c *PartedCacher) DeleteAll(keys []any) {
	c.ActiveCacher.DeleteAll(keys)
	c.LazyCacher.DeleteAll(keys)
}

func (c *PartedCacher) DeleteAllFn(fn func(key any, value *CacheItem) bool) {
	c.ActiveCacher.DeleteAllFn(fn)
	c.LazyCacher.DeleteAllFn(fn)
}

func (c *PartedCacher) DeleteLazy(key any) {
	c.LazyCacher.Delete(key)
}

func (c *PartedCacher) DeleteLazyAll(keys []any) {
	c.LazyCacher.DeleteAll(keys)
}

func (c *PartedCacher) IsFull() bool {
	return c.ActiveCacher.IsFull() && c.LazyCacher.IsFull()
}

func (c *PartedCacher) Clear() {
	c.ActiveCacher.Clear()
	c.LazyCacher.Clear()
}

func (c *PartedCacher) Len() int {
	return c.ActiveCacher.Len() + c.LazyCacher.Len()
}

func (c *PartedCacher) Keys() []any {
	return append(c.ActiveCacher.Keys(), c.LazyCacher.Keys()...)
}

func (c *PartedCacher) Values() []any {
	return append(c.ActiveCacher.Values(), c.LazyCacher.Values()...)
}

func (c *PartedCacher) Map() map[any]any {
	aMap := c.ActiveCacher.Map()
	lMap := c.LazyCacher.Map()
	for k := range lMap {
		aMap[k] = lMap[k]
	}
	return aMap
}

func (c *PartedCacher) PartedMap() map[string]map[any]any {
	aMap := c.ActiveCacher.Map()
	lMap := c.LazyCacher.Map()
	m := map[string]map[any]any{"active": {}, "lazy": {}}
	for k := range lMap {
		m["lazy"][k] = lMap[k]
	}
	for k := range aMap {
		m["active"][k] = aMap[k]
	}
	return m
}

type partedList struct {
	Active []any
	Lazy   []any
}

func (c *PartedCacher) PartedKeys() *partedList {
	return &partedList{
		Active: c.ActiveCacher.Keys(),
		Lazy:   c.LazyCacher.Keys(),
	}
}

func (c *PartedCacher) PartedValues() *partedList {
	return &partedList{
		Active: c.ActiveCacher.Values(),
		Lazy:   c.LazyCacher.Values(),
	}
}
