// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabin

import (
	"bytes"
	"io"
	"math/rand"
	"reflect"
	"testing"
)

func TestChunker(t *testing.T) {
	const (
		min = 128
		max = 4 << 10
	)

	nTests := 100
	if testing.Short() {
		nTests = 5
	}

	totalLen, numLen := 0, 0
	for nTest := 0; nTest < nTests; nTest++ {
		var l1, l2 []int
		rg := rand.New(rand.NewSource(int64(nTest)))
		data := make([]byte, 128<<10)
		rg.Read(data)
		tab := NewTable(Poly64, 64)

		// Chunk data using the Chunker.
		c := NewChunker(tab, bytes.NewReader(data), min, max)
		for {
			length, err := c.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Fatal("unexpected error", err)
			}
			l1 = append(l1, length)
		}

		// Chunk data using the obvious, slow, non-streaming
		// implementation.
		h := New(tab)
		clen := 0
		for _, b := range data {
			h.Write([]byte{b})
			clen++
			if (clen >= min && h.Sum64()&(0x1FFF) == (0)) ||
				clen == max {
				l2, clen = append(l2, clen), 0
			}
		}
		l2 = append(l2, clen)

		// Compare the results.
		if !reflect.DeepEqual(l1, l2) {
			t.Errorf("bad chunk lengths:\n want: %v\n got:  %v", l2, l1)
			continue
		}

		for _, l := range l1[:len(l1)-1] {
			totalLen += l
			numLen++
		}
	}

}

func BenchmarkChunker(b *testing.B) {
	const (
		min = 128
		max = 4 << 10
	)

	rg := rand.New(rand.NewSource(42))
	data := make([]byte, 1<<20)
	rg.Read(data)
	b.SetBytes(int64(len(data)))
	tab := NewTable(Poly64, 64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c := NewChunker(tab, bytes.NewReader(data), min, max)
		for {
			_, err := c.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				b.Fatal("unexpected error", err)
			}
		}
	}
}
