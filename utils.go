package golrucacher

import (
	"errors"
	"time"
)

type randDependTime struct {
	unixMicro int64
	decimal   float64
}

var (
	randDT *randDependTime = nil
)

const (
	randDTDecmical = 0.00001
)

// 根据时间获取 [0, 1) 之间的随机小数
// 该函数的目的是为了在同一微秒内，获取的随机数越晚，越大
//
// 要求：
// 1.（在同一微秒内）越晚调用函数，获取的随机数越大
// 2. 不同纳秒内，不需要在意大小
func _getRandDecmicalDependTime(unixMicro int64) (float64, error) {
	if randDT == nil {
		randDT = &randDependTime{
			unixMicro: unixMicro,
			decimal:   randDTDecmical,
		}
		return randDTDecmical, nil
	}
	if randDT.unixMicro != unixMicro {
		randDT.unixMicro = unixMicro
		randDT.decimal = randDTDecmical
		return randDTDecmical, nil
	}
	randDT.decimal += randDTDecmical
	if randDT.decimal >= 1 {
		return 0, errors.New("randDT.decimal >= 1")
	}
	return randDT.decimal, nil
}

func getTime() float64 {
	unixMicro := time.Now().UnixMicro()
	decmical, err := _getRandDecmicalDependTime(unixMicro)
	if err != nil {
		panic(err)
	}
	return float64(unixMicro) + decmical
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
