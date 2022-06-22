## Go LRU Cacher
`LRU` 算法缓存器。

### 用法
```go
package main

func main() {
    cacher := NewCacher(10)
    // 将键值对放入缓存器
    cacher.Update("foo", "bar")
    cacher.Update("foo2", "bar2")
    // 获取键值对
    cacher.Get("foo2") // "bar2"
    // 获取长度
    cacher.Len() // 2
    // 删除键值对
    cacher.Delete("foo")
    // 获取所有值
    cacher.All() // []string{"bar2"}
}
```