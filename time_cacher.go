package golrucacher

import "time"

// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
func NewDurationCacher(maxLength int, checkDuration time.Duration, fn func(key any, item *CacheItem) bool) *Cacher {
	c := NewCacher(maxLength)
	for range time.Tick(checkDuration) {
		go c.DeleteAllFn(fn)
	}
	return c
}

// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
func NewElapsedCacher(maxLength int, checkDuration, elapsedDuration time.Duration) *Cacher {
	elapsedUnixNano := elapsedDuration.Nanoseconds()
	c := NewDurationCacher(maxLength, checkDuration, func(key any, item *CacheItem) bool {
		return _unixNano()-item.LastTime > elapsedUnixNano
	})
	return c
}

func NewPartedDurationCacher(maxLength int, activeRate float64, checkDuration time.Duration, fn func(key any, item *CacheItem) bool) *PartedCacher {
	c := NewPartedCacher(maxLength, activeRate)
	for range time.Tick(checkDuration) {
		go c.DeleteAllFn(fn)
	}
	return c
}

func NewPartedElapsedCacher(maxLength int, activeRate float64, checkDuration, elapsedDuration time.Duration) *PartedCacher {
	elapsedUnixNano := elapsedDuration.Nanoseconds()
	c := NewPartedDurationCacher(maxLength, activeRate, checkDuration, func(key any, item *CacheItem) bool {
		return _unixNano()-item.LastTime > elapsedUnixNano
	})
	return c
}
