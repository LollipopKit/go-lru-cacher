package cacher

import "time"

type timeCacher struct {
	cacher *cacher
	fn func(key any, item *cacheItem) bool
	checkDuration time.Duration
	ticker *time.Ticker
}

// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
func NewTimeCacher(maxLength int, checkDuration time.Duration) *timeCacher {
	c := NewCacher(maxLength)
	return &timeCacher{
		cacher: c,
		checkDuration: checkDuration,
	}
}

func (tc *timeCacher) SetFn(fn func(key any, item *cacheItem) bool) {
	tc.fn = fn
	if tc.ticker != nil {
		tc.ticker.Stop()
	}
	tc.ticker = time.NewTicker(tc.checkDuration)
	for range tc.ticker.C {
		tc.cacher.DeleteAll(tc.fn)
	}
}

// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
func NewElapsedCacher(maxLength int, checkDuration, elapsedDuration time.Duration) *cacher {
	c := NewTimeCacher(maxLength, checkDuration)
	elapsedUnixNano := elapsedDuration.Nanoseconds()
	c.SetFn(func(key any, item *cacheItem) bool {
		return _unixNano() - item.lastTime > elapsedUnixNano
	})
	return c.cacher
}
