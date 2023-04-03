package golrucacher

import (
	"time"
)

func unixNano() int64 {
	return time.Now().UnixNano()
}

func calcMaxLength(maxLength int, activeRate float64) (a int, l int) {
	div := float64(maxLength) * activeRate
	a = int(div)
	if div-float64(a) >= 0.5 {
		a++
	}
	l = maxLength - a
	return
}
