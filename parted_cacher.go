package golrucacher

type PartedCacher struct {
	// 如果有部分数据一直被读取，可以使用 `lazy / active` 分区的缓存器，这样可以提高缓存命中率。
	// ### 逻辑
	// - 先把 `active` 写满，再写 `lazy`
	// - `active` 写满时写入，先将` active` 的最近、使用最多的缓存项“移动”到 `lazy`，再写入 `active`
	// - `active` 写满时写入，`lazy` 写满，删除 `lazy` 中最早、使用最少的缓存项，移动 `active` 的最近、使用最多的缓存项到 `lazy`，再写入 `active`

	active     *Cacher // “删”时优先处理该部分
	lazy       *Cacher // “查”时优先处理该部分
	activeRate float64 // active 部分的占比
}

// [activeRate]: 0 < activeRate < 1, activeRate = activeLen / maxLength
func NewPartedCacher(maxLength int, activeRate float64) *PartedCacher {
	if activeRate <= 0 || activeRate >= 1 {
		panic("lazyRate must be in (0, 1)")
	}
	a, l := _calcMaxLength(maxLength, activeRate)

	return &PartedCacher{
		active: NewCacher(a),
		lazy:   NewCacher(l),
	}
}

func (c *PartedCacher) Set(key, value any) {
	if c.lazy.caches[key] != nil {
		c.lazy.Set(key, value)
		return
	}
	if !c.active.IsFull() {
		c.active.Set(key, value)
		return
	}

	// active满：
	// 1、删除lazy中最早、使用最少的缓存项
	// 2、移动active的最近、使用最多的缓存项到lazy
	// 3、将新增项目添加到active
	delKeyInActive, lTime, lTimes := c.active.Laziest()
	v, ok := c.active.Get(delKeyInActive)
	if !c.lazy.IsFull() {
		if ok {
			c.lazy.Set(delKeyInActive, v)
			c.active.Delete(delKeyInActive)
			c.Set(key, value)
		} else {
			c.active.Set(key, value)
		}
		return
	}

	// lazy满：
	// 1、删除lazy中最早、使用最少的缓存项
	// 2、移动active的最近、使用最多的缓存项到lazy
	// 3、将新增项目添加到active
	delKeyInLazy, eTime, eTimes := c.lazy.Activest()
	if eTime <= lTime || eTimes <= lTimes { // FIFO：先进先出，所以包含等于
		c.lazy.Delete(delKeyInLazy)
		c.lazy.Set(delKeyInActive, v)
		c.active.Delete(delKeyInActive)
		c.Set(key, value)
		return
	}

	// lazy满，但lazy中最早、使用最少的缓存项不是active中最近、使用最多的缓存项
	// 所以，只需将新增项目添加到active
	c.active.Set(key, value)
}

func (c *PartedCacher) Get(key any) (any, bool) {
	v, ok := c.lazy.Get(key)
	if ok {
		return v, ok
	}

	return c.active.Get(key)
}

func (c *PartedCacher) Delete(key any) {
	c.active.Delete(key)
	c.lazy.Delete(key)
}

func (c *PartedCacher) DeleteAll(keys []any) {
	c.active.DeleteAll(keys)
	c.lazy.DeleteAll(keys)
}

func (c *PartedCacher) DeleteAllFn(fn func(key any, value *CacheItem) bool) {
	c.active.DeleteAllFn(fn)
	c.lazy.DeleteAllFn(fn)
}

func (c *PartedCacher) DeleteLazy(key any) {
	c.lazy.Delete(key)
}

func (c *PartedCacher) DeleteLazyAll(keys []any) {
	c.lazy.DeleteAll(keys)
}

func (c *PartedCacher) IsFull() bool {
	return c.active.IsFull() && c.lazy.IsFull()
}

func (c *PartedCacher) Clear() {
	c.active.Clear()
	c.lazy.Clear()
}

func (c *PartedCacher) Len() int {
	return c.active.Len() + c.lazy.Len()
}

func (c *PartedCacher) Keys() []any {
	return append(c.active.Keys(), c.lazy.Keys()...)
}

func (c *PartedCacher) Values() []any {
	return append(c.active.Values(), c.lazy.Values()...)
}

func (c *PartedCacher) Map() map[any]any {
	aMap := c.active.Map()
	lMap := c.lazy.Map()
	for k := range lMap {
		aMap[k] = lMap[k]
	}
	return aMap
}
