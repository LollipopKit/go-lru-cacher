package cacher

import "time"

func _unixNano() int64 {
	return time.Now().UnixNano()
}

func _calcMaxLength(maxLength int) (int, int) {
	r := int(float64(maxLength) * initRecentAndLazyRate)
	return r, maxLength - r
}
