## Go LRU Cacher

### 普通缓存器
```go
package main

import (
    glc "github.com/lollipopkit/go-lru-cacher"
)

func main() {
    // 创建一个最大容量为 10 的 LRU 缓存器
    cacher := glc.NewCacher(10)
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

### 分区缓存器
如果有部分数据一直被读取，可以使用 `lazy / active` 分区的缓存器，这样可以提高缓存命中率。

### 概念
- `active` 最近的，“删”时优先处理该部分
- `lazy` 低优先级的，“查”时优先处理该部分
- `activeRate` 为 `active` 部分的长度与 `lazy + active` 部分的长度的比值。目前该值不会变化，只能在初始化时设置。后继可能会随着读写操作动态改变。

### 逻辑
- 先把 `active` 写满，再写 `lazy`
- `active` 写满时写入，先将 `active` 的最近、使用最多的缓存项“移动”到 `lazy`，再写入 `active`
- `active` 写满时写入，`lazy` 写满，删除 `lazy` 中最早、使用最少的缓存项，移动 `active` 的最近、使用最多的缓存项到 `lazy`，再写入 `active`

### 使用
```go
// lazy + active 最大长度 = 10， lazy / lazy + active = 0.8
// active 区最多储存 8 个，lazy最多储存 2 个
NewPartedCacher(10, 0.8)
```
其他接口与上相同

### 超时缓存器
```go
// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
// 例：每过一小时，清理超过一小时未访问的缓存项
NewDurationCacher(10, time.Hour, func(key string, item *cacheItem) {
    if item.lastTime - time.Now().Nanoseconds() > 1000 * 1000 * 1000 * 60 * 60 {
        return true
    }
})
// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
// 例：每过十分钟，清理超过一小时未访问的缓存项
NewElapsedCacher(10, 10 * time.Minute, time.Hour)
```
其他接口与上相同