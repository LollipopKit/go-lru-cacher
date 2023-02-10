package golrucacher_test

import (
	"testing"

	glc "git.lolli.tech/lollipopkit/go-lru-cacher"
)

const (
	maxLength = 10
	activeRate  = 0.8
)

var (
	cacher       = glc.NewCacher(maxLength)
	partedCacher = glc.NewPartedCacher(maxLength, activeRate)
)

/*

Test

*/

func TestCacher(t *testing.T) {
	cacher.Set("key", "value")
	cacher.Set("key2", "value2")
	if cacher.Len() != 2 {
		t.Error("cacher.Len() != 2")
	}
	if v, ok := cacher.Get("key"); v.(string) != "value" || !ok {
		t.Error("cacher.Get(\"key\") != \"value\"")
	}

	cacher.Delete("key")
	if cacher.Len() != 1 {
		t.Error("cacher.Len() != 1")
	}
	if _, ok := cacher.Get("key"); ok {
		t.Error("cacher.Get(\"key\") != nil")
	}

	vs := cacher.Values()
	if len(vs) != 1 || vs[0] != "value2" {
		t.Error("cacher.Values() != [\"value2\"]")
	}

	ks := cacher.Keys()
	if len(ks) != 1 || ks[0] != "key2" {
		t.Error("cacher.Keys() != [\"key2\"]")
	}
	cacher.Clear()
	if cacher.Len() != 0 {
		t.Error("cacher.Len() != 0")
	}

	for i := 0; i < maxLength+2; i++ {
		cacher.Set(i, i)
	}
	if cacher.Len() != maxLength {
		t.Error("cacher.Len() != maxLength")
	}
	cacher.Get(2)
	for i := 0; i < maxLength-1; i++ {
		k, _, _ := cacher.Coldest()
		cacher.Delete(k)
	}
	if two, ok := cacher.Get(2); two != 2 && !ok {
		t.Log(two, ok)
		t.Error("cacher.Get(2) != 2")
	}
}

func TestPartedCacher(t *testing.T) {
	for i := 0; i < maxLength; i++ {
		partedCacher.Set(i, i)
	}

	for i := 0; i < maxLength * activeRate; i++ {
		partedCacher.Set(i, i + 100)
	}

	if v, ok := partedCacher.Get(8); v != 8 || !ok {
		t.Error("partedCacher.Get(8) != 8")
	}
	if v, ok := partedCacher.Get(9); v != 9 || !ok {
		t.Error("partedCacher.Get(9) != 9")
	}
}

/*

Bench

*/

func bench(item any, b *testing.B) {
	cacher.Clear()
	for i := 0; i < b.N; i++ {
		cacher.Set(i, item)
	}
	for i := 0; i < b.N; i++ {
		cacher.Get(i)
	}
}

func BenchmarkInt(b *testing.B) {
	bench(1, b)
}

func BenchmarkStruct(b *testing.B) {
	t := struct {
		tt []any
	}{
		tt: []any{glc.NewCacher(100), glc.NewPartedCacher(10, 0.5)},
	}
	bench(t, b)
}
