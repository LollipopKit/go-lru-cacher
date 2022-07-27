package cacher_test

import (
	"testing"
	glc "git.lolli.tech/lollipopkit/go_lru_cacher"
)

const (
	maxLength = 100
)

var (
	cacher = glc.NewCacher(maxLength)
)

func Test(t *testing.T) {
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

	for i := 0; i < maxLength + 2; i++ {
		cacher.Set(i, i)
	}
	if cacher.Len() != maxLength {
		t.Error("cacher.Len() != maxLength")
	}
	for i := 2; i < maxLength + 2; i++ {
		if v, ok := cacher.Get(i); !ok {
			t.Log(i, v)
			t.Error("cacher.Get(i) not ok")
		}
	}
}