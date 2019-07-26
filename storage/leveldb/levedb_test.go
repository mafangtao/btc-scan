package leveldb

import (
	"fmt"
	"testing"
)

//go test -v -run='TestSetGet'
func TestSetGet(t *testing.T) {

	s, err := NewLevelDbStorage("./ldb")
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := s.Set(fmt.Sprintf("key-%02d", i), fmt.Sprintf("%d", i))
		if err != nil {
			t.Error(err)
			return
		}
	}

	key := fmt.Sprintf("key-%02d", 5)
	val, err := s.Get(key)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("key:%s, val:%s", key, val)
	}
}

//go test -v -run='TestScan'
func TestScan(t *testing.T) {

	s, err := NewLevelDbStorage("./ldb")
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := s.Set(fmt.Sprintf("key-%02d", i), fmt.Sprintf("%d", i))
		if err != nil {
			t.Error(err)
			return
		}
	}

	keyStart := fmt.Sprintf("key-%02d", 0)
	keyEnd := fmt.Sprintf("key-%02d", 9)

	m , err := s.Scan(keyStart, keyEnd, 10)

	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("scan:%v", m)
	}

	keyStart = fmt.Sprintf("key-")
	m , err = s.Scan(keyStart, keyEnd, 10)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("scan:%v", m)
	}

	keyStart = fmt.Sprintf("key-")
	m , err = s.Scan(keyStart, keyEnd, 5)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("scan:%v", m)
	}

	keyStart = fmt.Sprintf("key-%02d", 2)
	m , err = s.Scan(keyStart, keyEnd, 3)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("scan:%v", m)
	}
}



//go test -v -run='TestRScan'
func TestRScan(t *testing.T) {

	s, err := NewLevelDbStorage("./ldb")
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := s.Set(fmt.Sprintf("key-%02d", i), fmt.Sprintf("%d", i))
		if err != nil {
			t.Error(err)
			return
		}
	}

	keyStart := fmt.Sprintf("key-%02d", 10)
	keyEnd := fmt.Sprintf("key-%02d", 0)

	m , err := s.RScan(keyStart, keyEnd, 10)

	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("rscan:%v", m)
	}

	keyStart = fmt.Sprintf("key-%02d", 7)
	keyEnd = fmt.Sprintf("key-%02d", 3)
	m , err = s.RScan(keyStart, keyEnd, 10)

	if err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("rscan:%v", m)
	}
}

