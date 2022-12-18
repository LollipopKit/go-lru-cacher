## Go LRU Cacher

### 普通缓存器
```go
package main

func main() {
    // 创建一个最大容量为 10 的 LRU 缓存器
    cacher := NewCacher(10)
    // 将键值对放入缓存器
    cacher.Set("foo", "bar")
    cacher.Set("foo2", "bar2")
    // 获取键值对
    cacher.Get("foo2") // "bar2"
    // 获取长度
    cacher.Len() // 2
    // 删除键值对
    cacher.Delete("foo")
    // 获取所有值
    cacher.Values() // []interface{}{"bar2"}
    // 获取所有键
    cacher.Keys() // []string{"foo2"}
    // 清空缓存器
    cacher.Clear()
}
```

### 冷热分区缓存器
如果有一些数据一直被读写，可以使用冷热分区的缓存器，这样可以提高缓存命中率。
```go
NewPartedCacher(10)
```
其他接口与上相同

### 超时检测缓存器
```go
// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
// 例：每过一小时，清理超过一小时未访问的缓存项
NewTimeCacher(10, time.Hour, func(key string, item *cacheItem) {
    if item.lastTime - time.Now().Nanoseconds() > 1000 * 1000 * 1000 * 60 * 60 {
        return true
    }
})
// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
// 例：每过十分钟，清理超过一小时未访问的缓存项
NewElapsedCacher(10, 10 * time.Minute, time.Hour)
```
其他接口与上相同