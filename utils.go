package cacher

import "time"

func _unixNano() int64 {
	return time.Now().UnixNano()
}
