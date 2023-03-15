package golrucacher

import (
	"fmt"
	"time"
)

type termColor string

const (
	red    termColor = "\033[31m"
	green  termColor = "\033[32m"
	yellow termColor = "\033[33m"
	cyan   termColor = "\033[36m"
	reset  termColor = "\033[0m"
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

func log(color termColor, subj string, msg any) {
	fmt.Printf("%s[%s]%s: %v\n", color, subj, reset, msg)
}
