package cacher

import "time"

func _unixNano() int64 {
	return time.Now().UnixNano()
}

func _calcMaxLength(maxLength int) (int, int) {
	if maxLength <= 0 {
		panic("maxLength must be greater than 0")
	}

	r := int(float64(maxLength) * initRecentAndLazyRate)
	return r, maxLength - r
}
