package lcdb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/mihongtech/linkchain/common/lcdb"
	"github.com/mihongtech/linkchain/common/math"
	"github.com/mihongtech/linkchain/unittest"
	"time"
)

func newTestLDB() (*lcdb.LDBDatabase, func()) {
	dirname, err := ioutil.TempDir(os.TempDir(), "lcdb_test_")
	if err != nil {
		panic("failed to create test file: " + err.Error())
	}
	db, err := lcdb.NewLDBDatabase(dirname, 0, 0)
	if err != nil {
		panic("failed to create test database: " + err.Error())
	}

	return db, func() {
		db.Close()
		os.RemoveAll(dirname)
	}
}

var test_values = []string{"", "a", "1251", "\x00123\x00"}

func TestLDB_PutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	testPutGet(db, t)
}

func TestMemoryDB_PutGet(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	testPutGet(db, t)
}

func testPutGet(db lcdb.Database, t *testing.T) {
	t.Parallel()

	for _, v := range test_values {
		err := db.Put([]byte(v), []byte(v))
		if err != nil {
			t.Fatalf("put failed: %v", err)
		}
	}

	for _, v := range test_values {
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte(v)) {
			t.Fatalf("get returned wrong result, got %q expected %q", string(data), v)
		}
	}

	for _, v := range test_values {
		err := db.Put([]byte(v), []byte("?"))
		if err != nil {
			t.Fatalf("put override failed: %v", err)
		}
	}

	for _, v := range test_values {
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte("?")) {
			t.Fatalf("get returned wrong result, got %q expected ?", string(data))
		}
	}

	for _, v := range test_values {
		orig, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		orig[0] = byte(0xff)
		data, err := db.Get([]byte(v))
		if err != nil {
			t.Fatalf("get failed: %v", err)
		}
		if !bytes.Equal(data, []byte("?")) {
			t.Fatalf("get returned wrong result, got %q expected ?", string(data))
		}
	}

	for _, v := range test_values {
		err := db.Delete([]byte(v))
		if err != nil {
			t.Fatalf("delete %q failed: %v", v, err)
		}
	}

	for _, v := range test_values {
		_, err := db.Get([]byte(v))
		if err == nil {
			t.Fatalf("got deleted value %q", v)
		}
	}
}

func TestLDB_ParallelPutGet(t *testing.T) {
	db, remove := newTestLDB()
	defer remove()
	testParallelPutGet(db, t)
}

func TestMemoryDB_ParallelPutGet(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	testParallelPutGet(db, t)
}

func testParallelPutGet(db lcdb.Database, t *testing.T) {
	const n = 8
	var pending sync.WaitGroup

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			err := db.Put([]byte(key), []byte("v"+key))
			if err != nil {
				panic("put failed: " + err.Error())
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			data, err := db.Get([]byte(key))
			if err != nil {
				panic("get failed: " + err.Error())
			}
			if !bytes.Equal(data, []byte("v"+key)) {
				panic(fmt.Sprintf("get failed, got %q expected %q", []byte(data), []byte("v"+key)))
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			err := db.Delete([]byte(key))
			if err != nil {
				panic("delete failed: " + err.Error())
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()

	pending.Add(n)
	for i := 0; i < n; i++ {
		go func(key string) {
			defer pending.Done()
			_, err := db.Get([]byte(key))
			if err == nil {
				panic("get succeeded")
			}
		}(strconv.Itoa(i))
	}
	pending.Wait()
}

func TestMaxSizeValue(t *testing.T) {
	db, _ := lcdb.NewMemDatabase()
	accountID := math.DoubleHashH([]byte("lifei"))
	accountIDByte := accountID.CloneBytes()
	value := make([]byte, 10000)
	key := []byte("lifei")
	for i := 0; i < 1000000; i++ {
		value = append(value, accountIDByte...)
	}
	valueHash := math.DoubleHashH(value)
	time1 := timeGet(t)
	//put
	if err := db.Put(key, value); err != nil {
		t.Error("db put failed   ", err)
	}
	t.Log("value size (M)", len(value)/(1024*1024))
	time2 := timeGet(t)
	t.Log("put 1000000 accountID waste time(ms)", (time2-time1)/1e6)

	value, err := db.Get(key)
	if err != nil {
		t.Error("db get failed   ", err)
	}
	time3 := timeGet(t)
	t.Log("get 1000000 accountID waste time(ms)", (time3-time2)/1e6)
	valueHash1 := math.DoubleHashH(value)
	unittest.Equal(t, valueHash1.IsEqual(&valueHash), true)

	time4 := timeGet(t)
	value = append(value[:10000], value[10257:]...)
	if err := db.Put(key, value); err != nil {
		t.Error("db update failed   ", err)
	}
	time5 := timeGet(t)
	t.Log("update 1000000 accountID waste time(ms)", (time5-time4)/1e6)
	valueHash2 := math.DoubleHashH(value)
	unittest.NotEqual(t, valueHash2.IsEqual(&valueHash), true)

	time6 := timeGet(t)
	db.Delete(key)
	time7 := timeGet(t)
	t.Log("delete 1000000 accountID waste time(ms)", (time7-time6)/1e6)
}

func timeGet(t *testing.T) int64 {
	return time.Now().UnixNano()
}
