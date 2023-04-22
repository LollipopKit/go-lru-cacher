package golrucacher_test

import (
	"fmt"
	"testing"
	"time"

	glc "github.com/lollipopkit/go-lru-cacher"
)

const (
	maxLength  = 10
	activeRate = 0.8
)

var (
	cacher        = glc.NewCacher[test[string]](maxLength)
	partedCacher  = glc.NewPartedCacher[test[int]](maxLength, activeRate)
	elapsedCacher = glc.NewElapsedCacher[test[int]](10, time.Second, time.Second)
)

type test[T any] struct {
	Value T
}

func genTest[T any](v T) *test[T] {
	return &test[T]{Value: v}
}

/*

Test

*/

func TestCacher(t *testing.T) {
	cacher.Set("key", genTest("value"))
	cacher.Set("key2", genTest("value2"))
	if cacher.Len() != 2 {
		t.Error("cacher.Len() != 2")
	}
	if v, ok := cacher.Get("key"); v.Value != "value" || !ok {
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
	if len(vs) != 1 || vs[0].Value != "value2" {
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
		cacher.Set(i, genTest(fmt.Sprintf("%d", i)))
	}
	if cacher.Len() != maxLength {
		t.Error("cacher.Len() != maxLength")
	}
	cacher.Get(3)
	for i := 0; i < maxLength-1; i++ {
		k, _, _ := cacher.Activest()
		cacher.Delete(k)
	}
	if two, ok := cacher.Get(3); !ok || two.Value != "3" {
		t.Log(two, ok)
		t.Error("cacher.Get(3) != 3")
	}
}

func TestPartedCacher(t *testing.T) {
	for i := 0; i < maxLength; i++ {
		partedCacher.Set(i, genTest(i))
	}

	for i := 0; i < maxLength*activeRate; i++ {
		partedCacher.Set(i, genTest(i+100))
	}

	if v, ok := partedCacher.Get(8); v.Value != 8 || !ok {
		t.Error("partedCacher.Get(8) != 8")
	}
	if v, ok := partedCacher.Get(9); v.Value != 9 || !ok {
		t.Error("partedCacher.Get(9) != 9")
	}
}

func TestElapsedCacher(t *testing.T) {
	elapsedCacher.Set(1, genTest(1))
	time.Sleep(time.Second * 2)
	if _, ok := elapsedCacher.Get(1); ok {
		t.Error("elapsedCacher.Get(1) != nil")
	}
}
