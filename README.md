## Go LRU Cacher
`LRU` 算法缓存器。

### 用法
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