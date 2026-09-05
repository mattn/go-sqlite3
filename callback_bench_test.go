// Copyright (C) 2019 Yasuhiro Matsumoto <mattn.jp@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//go:build cgo
// +build cgo

package sqlite3

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func BenchmarkHandleLookupParallel(b *testing.B) {
	d := SQLiteDriver{}
	conn, err := d.Open(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()
	c := conn.(*SQLiteConn)

	handle := newHandle(c, func() {})

	benchmarkHandleLookupParallel(b, func() any {
		return lookupHandle(handle)
	})
}

func BenchmarkHandleLookupBeforeAfter(b *testing.B) {
	value := handleVal{val: func() {}}
	handle := unsafe.Pointer(&value)

	before := mutexHandleTable{vals: map[unsafe.Pointer]handleVal{handle: value}}
	after := atomicHandleTable{}
	after.vals.Store(map[unsafe.Pointer]handleVal{handle: value})

	b.Run("before_mutex", func(b *testing.B) {
		benchmarkHandleLookupParallel(b, func() any {
			return before.lookup(handle).val
		})
	})
	b.Run("after_atomic", func(b *testing.B) {
		benchmarkHandleLookupParallel(b, func() any {
			return after.lookup(handle).val
		})
	})
}

func benchmarkHandleLookupParallel(b *testing.B, lookup func() any) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if lookup() == nil {
				b.Fatal("lookup returned nil")
			}
		}
	})
}

type mutexHandleTable struct {
	mu   sync.Mutex
	vals map[unsafe.Pointer]handleVal
}

func (t *mutexHandleTable) lookup(handle unsafe.Pointer) handleVal {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.vals[handle]
}

type atomicHandleTable struct {
	vals atomic.Value
}

func (t *atomicHandleTable) lookup(handle unsafe.Pointer) handleVal {
	m, _ := t.vals.Load().(map[unsafe.Pointer]handleVal)
	return m[handle]
}
