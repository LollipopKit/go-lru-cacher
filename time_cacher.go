package golrucacher

import "time"

// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
func NewDurationCacher[T any](maxLength int, checkDuration time.Duration, fn func(key any, item *CacheItem[T]) bool) *cacher[T] {
	c := NewCacher[T](maxLength)
	go func() {
		for range time.Tick(checkDuration) {
			c.DeleteAllFn(fn)
		}
	}()
	return c
}

// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
func NewElapsedCacher[T any](maxLength int, checkDuration, elapsedDuration time.Duration) *cacher[T] {
	elapsedUnixNano := float64(elapsedDuration.Microseconds())
	c := NewDurationCacher(maxLength, checkDuration, func(key any, item *CacheItem[T]) bool {
		return getTime()-item.LastTime > elapsedUnixNano
	})
	return c
}

func NewPartedDurationCacher[T any](maxLength int, activeRate float64, checkDuration time.Duration, fn func(key any, item *CacheItem[T]) bool) *PartedCacher[T] {
	c := NewPartedCacher[T](maxLength, activeRate)
	go func() {
		for range time.Tick(checkDuration) {
			c.DeleteAllFn(fn)
		}
	}()
	return c
}

func NewPartedElapsedCacher[T any](maxLength int, activeRate float64, checkDuration, elapsedDuration time.Duration) *PartedCacher[T] {
	elapsedUnixNano := float64(elapsedDuration.Microseconds())
	c := NewPartedDurationCacher(maxLength, activeRate, checkDuration, func(key any, item *CacheItem[T]) bool {
		return getTime()-item.LastTime > elapsedUnixNano
	})
	return c
}
