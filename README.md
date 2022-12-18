## Go LRU Cacher
`LRU` 算法缓存器。

### 用法
#### 普通LRU缓存器
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

#### 冷热分区LRU缓存器
如果有一些数据一直被读写，可以使用冷热分区的缓存器，这样可以提高缓存命中率。
```go
NewPartedCacher(10)
```
其他接口与上相同

#### 自动过期LRU缓存器
```go
// 返回一个缓存器，每经过 duration 调用一次 fn 自定义清理缓存项
NewTimeCacher()
// 每过 checkDuration 检查一次，间隔超过 elapsedDuration 的缓存项将被清理
NewElapsedCacher(10, 10 * time.Minute, time.Hour)
```