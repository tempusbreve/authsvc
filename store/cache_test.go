package store

import (
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

var (
	past    = func() time.Time { return time.Date(2008, time.January, 3, 14, 59, 59, 0, time.UTC) }
	present = func() time.Time { return time.Date(2018, time.April, 26, 17, 00, 00, 0, time.UTC) }
	future  = func() time.Time { return time.Date(2028, time.November, 3, 21, 29, 39, 0, time.UTC) }
)

type factory func(func() time.Time) Cache
type testFn func(string, Cache, testCase, *testing.T)

var (
	memoryFactory factory = func(now func() time.Time) Cache { return &memory{now: now, data: map[string]*cacheValue{}} }
	boltFactory   factory = func(now func() time.Time) Cache {
		name, err := tempFile()
		if err != nil {
			panic(err)
		}
		c, err := NewBoltDBCache(name)
		if err != nil {
			panic(err)
		}
		return c
	}
)

func TestMemoryCache(t *testing.T) {
	testCache(memoryFactory, t)
}

func TestBoltDBCache(t *testing.T) {
	testCache(boltFactory, t)
}

func testCache(fn factory, t *testing.T) {
	testCases := map[string]testCase{
		"one":   {key: "key", vv: "value"},
		"two":   {key: "key", vv: 42},
		"three": {key: "key", vv: false, expire: past(), experr: ErrExpired},
		"four":  {key: "key", vv: time.Now().UTC().Round(0), expire: future()},
	}

	testFunctions := []testFn{
		testPut,
		testGet,
		testPutUntil,
		testGetExpire,
		testPutUntilFuture,
		testGetFuture,
		testPutUntilPast,
		testGetPast,
	}

	for name, tc := range testCases {
		var cache Cache
		now := tc.now
		if now == nil {
			now = present
		}
		cache = fn(now)

		for _, fn := range testFunctions {
			fn(name, cache, tc, t)
		}
	}
}

type testCase struct {
	key    string
	vv     interface{}
	expire time.Time
	now    clockFn
	puterr error
	geterr error
	experr error
}

func testPut(name string, cache Cache, tc testCase, t *testing.T) {
	if err := cache.Put(tc.key, tc.vv); err != tc.puterr {
		t.Errorf("%q: Put() result expected %v, got %v", name, tc.puterr, err)
	}
}

func testGet(name string, cache Cache, tc testCase, t *testing.T) {
	v, err := cache.Get(tc.key)
	if err != tc.geterr {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.geterr, err)
	}
	if !reflect.DeepEqual(v, tc.vv) {
		t.Errorf("%q: Get() result value expected %v (%T), got %v (%T)", name, tc.vv, tc.vv, v, v)
	}

	v, err = cache.Get(tc.key + "other")
	if err != ErrNotFound {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.geterr, err)
	}
	if v != nil {
		t.Errorf("%q: Get() result value expected nil, got %v (%T)", name, v, v)
	}
}

func testPutUntil(name string, cache Cache, tc testCase, t *testing.T) {
	if err := cache.PutUntil(tc.expire, tc.key, tc.vv); err != tc.puterr {
		t.Errorf("%q: PutUntil() result expected %v, got %v", name, tc.puterr, err)
	}
}

func testGetExpire(name string, cache Cache, tc testCase, t *testing.T) {
	v, err := cache.Get(tc.key)
	if err != tc.experr {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.experr, err)
	}
	if !reflect.DeepEqual(v, tc.vv) {
		t.Errorf("%q: Get() result value expected %v (%T), got %v (%T)", name, tc.vv, tc.vv, v, v)
	}
}

func testPutUntilFuture(name string, cache Cache, tc testCase, t *testing.T) {
	if err := cache.PutUntil(future(), tc.key, tc.vv); err != tc.puterr {
		t.Errorf("%q: PutUntil() result expected %v, got %v", name, tc.puterr, err)
	}
}

func testGetFuture(name string, cache Cache, tc testCase, t *testing.T) {
	v, err := cache.Get(tc.key)
	if err != tc.geterr {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.geterr, err)
	}
	if !reflect.DeepEqual(v, tc.vv) {
		t.Errorf("%q: Get() result value expected %v (%T), got %v (%T)", name, tc.vv, tc.vv, v, v)
	}
}

func testPutUntilPast(name string, cache Cache, tc testCase, t *testing.T) {
	if err := cache.PutUntil(past(), tc.key, tc.vv); err != tc.puterr {
		t.Errorf("%q: PutUntil() result expected %v, got %v", name, tc.puterr, err)
	}
}

func testGetPast(name string, cache Cache, tc testCase, t *testing.T) {
	v, err := cache.Get(tc.key)
	if err != ErrExpired {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.geterr, err)
	}
	if !reflect.DeepEqual(v, tc.vv) {
		t.Errorf("%q: Get() result value expected %v (%T), got %v (%T)", name, tc.vv, tc.vv, v, v)
	}

	v, err = cache.Get(tc.key)
	if err != ErrNotFound {
		t.Errorf("%q: Get() result error expected %v, got %v", name, tc.geterr, err)
	}
	if v != nil {
		t.Errorf("%q: Get() result value expected nil, got %v (%T)", name, v, v)
	}
}

func tempFile() (string, error) {
	f, err := ioutil.TempFile("", "store_test")
	if err != nil {
		return "", err
	}
	name := f.Name()
	return name, f.Close()
}
