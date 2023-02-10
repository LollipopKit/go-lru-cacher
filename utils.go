package golrucacher

import "time"

func _unixNano() int64 {
	return time.Now().UnixNano()
}

func _calcMaxLength(maxLength int, activeRate float64) (int, int) {
	r := int(float64(maxLength) * activeRate)
	return r, maxLength - r
}
