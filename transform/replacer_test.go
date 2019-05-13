package transform_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"golang.org/x/text/transform"

	. "github.com/tenntenn/text/transform"
)

type history struct {
	src0, src1 []int
	dst0, dst1 []int
}

// ExampleReplaceAll is an example of ReplaceAll.
func ExampleReplaceTable() {
	t := ReplaceStringTable{
		"Hello", "Hi",
		"World", "Gophers",
	}
	r := transform.NewReader(strings.NewReader("Hello, World"), ReplaceAll(t))
	io.Copy(os.Stdout, r)
	// Output: Hi, Gophers
}

// TestReplace is a test for Replace.Transform.
func TestReplacer_Transform(t *testing.T) {
	data := []struct {
		// input
		old, new []byte
		dst, src []byte
		atEOF    bool

		// expected
		nDst, nSrc int
		hasErr     bool
		expected   []byte
	}{
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefgabcd"),
			atEOF:    true,
			nDst:     11,
			nSrc:     11,
			expected: []byte(`ABCdefgABCd`),
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefgabcd"),
			atEOF:    false,
			nDst:     11,
			nSrc:     11,
			expected: []byte(`ABCdefgABCd`),
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefgabca"),
			atEOF:    false,
			nDst:     10,
			nSrc:     11,
			expected: []byte(`ABCdefgABC`),
			hasErr:   true,
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefgabca"),
			atEOF:    true,
			nDst:     11,
			nSrc:     11,
			expected: []byte(`ABCdefgABCa`),
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefabca"),
			atEOF:    false,
			nDst:     9,
			nSrc:     10,
			expected: []byte(`ABCdefABC`),
			hasErr:   true,
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefabca"),
			atEOF:    true,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`ABCdefABCa`),
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 100),
			src:      []byte("abcdefgabc"),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`ABCdefgABC`),
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 2),
			src:      []byte(`abc`),
			atEOF:    false,
			nDst:     2,
			nSrc:     3,
			expected: []byte(`AB`),
			hasErr:   true,
		},
		{ // 8
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 2),
			src:      []byte(`abc`),
			atEOF:    false,
			nDst:     2,
			nSrc:     3,
			expected: []byte(`AB`),
			hasErr:   true,
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 2),
			src:      []byte(`ab`),
			atEOF:    false,
			nDst:     0,
			nSrc:     2,
			expected: []byte(``),
			hasErr:   true,
		},
		{
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			dst:      make([]byte, 2),
			src:      []byte(`ab`),
			atEOF:    true,
			nDst:     2,
			nSrc:     2,
			expected: []byte(`ab`),
		},
		// tests for nil and empty bytes
		{
			old:      nil,
			new:      nil,
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      nil,
			new:      []byte(`abc`),
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      []byte(`123`),
			new:      nil,
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     7,
			nSrc:     10,
			expected: []byte(`0456789`),
		},
		{
			old:      []byte(`12a`),
			new:      nil,
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      []byte{},
			new:      []byte{},
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      []byte{},
			new:      []byte(`abc`),
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      []byte(`123`),
			new:      []byte{},
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     7,
			nSrc:     10,
			expected: []byte(`0456789`),
		},
		{
			old:      []byte(`12a`),
			new:      []byte{},
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      []byte{},
			new:      nil,
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		{
			old:      nil,
			new:      []byte{},
			dst:      make([]byte, 100),
			src:      []byte(`0123456789`),
			atEOF:    false,
			nDst:     10,
			nSrc:     10,
			expected: []byte(`0123456789`),
		},
		// -- end of tests for nil and empty bytes
	}

	for i, d := range data {
		nDst, nSrc, err := NewReplacer(d.old, d.new, nil).Transform(d.dst, d.src, d.atEOF)
		switch {
		case d.hasErr && err == nil:
			t.Errorf("data[%d] must occur an error but not occured", i)
			continue
		case !d.hasErr && err != nil:
			t.Errorf("data[%d] must not occur an error but error occured: %v", i, err)
			continue
		}

		if nDst != d.nDst {
			t.Errorf("data[%d]'s expected nDst is %d but %d", i, d.nDst, nDst)
		} else if bytes.Compare(d.dst[:nDst], d.expected) != 0 {
			t.Errorf("data[%d]'s expected dst is %v but %v", i, d.expected, d.dst[:nDst])
		}

		if nSrc != d.nSrc {
			t.Errorf("data[%d]'s expected nSrc is %d but %d", i, d.nSrc, nSrc)
		}
	}
}

func TestReplacer_TransformFlow(t *testing.T) {
	type flow []struct {
		dst, src []byte
		atEOF    bool

		nSrc int
		out  []byte // nDst == len(out)
		err  error
	}
	cases := []struct {
		old, new []byte
		flow     flow
		history  *history
	}{
		{
			old: []byte(`abc`),
			new: []byte(`R`),
			flow: flow{
				{
					dst:   make([]byte, 5),
					src:   []byte(`abcde`),
					atEOF: false,
					out:   []byte(`Rde`),
					nSrc:  5,
					err:   nil,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`fabcd`),
					atEOF: true,
					out:   []byte(`fRd`),
					nSrc:  5,
					err:   nil,
				},
			},
			history: &history{
				src0: []int{0, 6},
				src1: []int{3, 9},
				dst0: []int{0, 4},
				dst1: []int{1, 5},
			},
		},
		{
			old: []byte(`abc`),
			new: []byte(`ABC`),
			flow: flow{
				{
					dst:   make([]byte, 2),
					src:   []byte(`ab`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  2,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 2),
					src:   []byte(`cd`),
					atEOF: true,
					out:   []byte(`AB`),
					nSrc:  1,
					err:   transform.ErrShortDst,
				},
				{
					dst:   make([]byte, 10),
					src:   []byte(`de`),
					atEOF: true,
					out:   []byte(`Cde`),
					nSrc:  2,
					err:   nil,
				},
			},
			history: &history{
				src0: []int{0},
				src1: []int{3},
				dst0: []int{0},
				dst1: []int{3},
			},
		},
		{
			old: []byte(`abc`),
			new: []byte(`A`),
			flow: flow{
				{
					dst:   make([]byte, 2),
					src:   []byte(`ab`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  2,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 2),
					src:   []byte(`cd`),
					atEOF: true,
					out:   []byte(`Ad`),
					nSrc:  2,
					err:   nil,
				},
			},
			history: &history{
				src0: []int{0},
				src1: []int{3},
				dst0: []int{0},
				dst1: []int{1},
			},
		},
		{
			old: []byte(`abcabcabc`),
			new: []byte(`R`),
			flow: flow{
				{
					dst:   make([]byte, 5),
					src:   []byte(`abcab`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  5,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`cabca`),
					atEOF: false,
					out:   []byte(`R`),
					nSrc:  5,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`bcabc`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  5,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`abcab`),
					atEOF: true,
					out:   []byte(`Rab`),
					nSrc:  5,
					err:   nil,
				},
			},
			history: &history{
				src0: []int{0, 9},
				src1: []int{9, 18},
				dst0: []int{0, 1},
				dst1: []int{1, 2},
			},
		},
		{
			old: []byte(`abcdefghi`),
			new: []byte(`R`),
			flow: flow{
				{
					dst:   make([]byte, 5),
					src:   []byte(`abcde`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  5,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`fghia`),
					atEOF: false,
					out:   []byte(`R`),
					nSrc:  5,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 5),
					src:   []byte(`bcdef`),
					atEOF: true,
					out:   []byte(`abcde`),
					nSrc:  4,
					err:   transform.ErrShortDst,
				},
			},
			history: &history{
				src0: []int{0},
				src1: []int{9},
				dst0: []int{0},
				dst1: []int{1},
			},
		},
		{
			old: []byte(`abcabcdef`),
			new: []byte(`R`),
			flow: flow{
				{
					dst:   make([]byte, 6),
					src:   []byte(`abcabc`),
					atEOF: false,
					out:   []byte(``),
					nSrc:  6,
					err:   transform.ErrShortSrc,
				},
				{
					dst:   make([]byte, 6),
					src:   []byte(`abcdef`),
					atEOF: false,
					out:   []byte(`abcR`),
					nSrc:  6,
					err:   nil,
				},
			},
			history: &history{
				src0: []int{3},
				src1: []int{12},
				dst0: []int{3},
				dst1: []int{4},
			},
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			history := NewReplaceHistory()
			r := NewReplacer(c.old, c.new, history)
			for _, f := range c.flow {
				nDst, nSrc, err := r.Transform(f.dst, f.src, f.atEOF)
				if nDst != len(f.out) {
					t.Errorf("the nDst is expected %d but %d", len(f.out), nDst)
				}
				if !bytes.Equal(f.dst[:nDst], f.out) {
					t.Errorf("the dst is expected %v but %v", f.out, f.dst[:nDst])
				}
				if nSrc != f.nSrc {
					t.Errorf("the nSrc is expected %d but %d", f.nSrc, nSrc)
				}
				if err != f.err {
					t.Errorf("the err is expected %v but %v", f.err, err)
				}
			}

			for j := range c.history.src0 {
				src0, src1, dst0, dst1 := history.At(j)
				if c.history.src0[j] != src0 {
					t.Errorf("data[%d]'s expected src0 of history[%d] is %d but %d", i, j, c.history.src0[j], src0)
				}
				if c.history.src1[j] != src1 {
					t.Errorf("data[%d]'s expected src1 of history[%d] is %d but %d", i, j, c.history.src1[j], src1)
				}
				if c.history.dst0[j] != dst0 {
					t.Errorf("data[%d]'s expected dst0 of history[%d] is %d but %d", i, j, c.history.dst0[j], dst0)
				}
				if c.history.dst1[j] != dst1 {
					t.Errorf("data[%d]'s expected dst1 of history[%d] is %d but %d", i, j, c.history.dst1[j], dst1)
				}
			}
		})
	}
}

// TestReplacerWithReader is a test for transform.Replace with transform.Reader.
func TestReplacerWithReader(t *testing.T) {
	data := []struct {
		// input
		r        io.Reader
		old, new []byte

		// expected
		expected []byte
		hasErr   bool
		history  *history
	}{
		{
			r:        strings.NewReader(`abcdefgabcd`),
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			expected: []byte(`ABCdefgABCd`),
			history: &history{
				src0: []int{0, 7},
				src1: []int{3, 10},
				dst0: []int{0, 7},
				dst1: []int{3, 10},
			},
		},
		{ // '🍺' and '🍻' are 4 bytes.
			r:        strings.NewReader(`Cheers!🍺`),
			old:      []byte(`🍺`),
			new:      []byte(`🍻`),
			expected: []byte(`Cheers!🍻`),
			history: &history{
				src0: []int{7},
				src1: []int{11},
				dst0: []int{7},
				dst1: []int{11},
			},
		},
		// Because transform.Reader uses 4096 bytes buffer, following tests confirm long data.
		{
			r:        strings.NewReader(strings.Repeat("*", 10000) + `abcdefgabcd`),
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			expected: append(bytes.Repeat([]byte("*"), 10000), []byte(`ABCdefgABCd`)...),
		},
		{
			r:        strings.NewReader(strings.Repeat("abc", 10000)),
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			expected: bytes.Repeat([]byte("ABC"), 10000),
		},
		{
			r:        strings.NewReader(strings.Repeat("abc", 10000) + `012`),
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			expected: append(bytes.Repeat([]byte("ABC"), 10000), []byte(`012`)...),
		},
		{
			r:        strings.NewReader(`aaaa` + strings.Repeat("abcdefg", 5000)),
			old:      []byte(`abc`),
			new:      []byte(`ABC`),
			expected: append([]byte(`aaaa`), bytes.Repeat([]byte("ABCdefg"), 5000)...),
		},
		{
			r:        strings.NewReader("aaabc"),
			old:      []byte(`abc`),
			new:      bytes.Repeat([]byte("A"), 5000),
			expected: append([]byte(`aa`), bytes.Repeat([]byte("A"), 5000)...),
		},
		{
			r:        strings.NewReader(strings.Repeat("abc", 10000)),
			old:      bytes.Repeat([]byte("abc"), 2000),
			new:      []byte(`REPLACED`),
			expected: bytes.Repeat([]byte("REPLACED"), 5),
		},
		{
			r:        strings.NewReader(strings.Repeat("abc", 1999) + "ab" + strings.Repeat("abc", 2000)),
			old:      bytes.Repeat([]byte("abc"), 2000),
			new:      []byte(`REPLACED`),
			expected: append(bytes.Repeat([]byte("abc"), 2000)[:5999], []byte("REPLACED")...),
		},
	}

	for i, d := range data {
		history := NewReplaceHistory()
		r := NewReplacer(d.old, d.new, history)
		actual, err := ioutil.ReadAll(transform.NewReader(d.r, r))
		switch {
		case d.hasErr && err == nil:
			t.Errorf("data[%d] must occur an error but not occured", i)
			continue
		case !d.hasErr && err != nil:
			t.Errorf("data[%d] must not occur an error but error occured: %v", i, err)
			continue
		}

		if bytes.Compare(actual, d.expected) != 0 {
			t.Errorf("data[%d]'s expected value is %v but %v", i, d.expected, actual)
		}

		if d.history != nil {
			for j := range d.history.src0 {
				src0, src1, dst0, dst1 := history.At(j)
				if d.history.src0[j] != src0 {
					t.Errorf("data[%d]'s expected src0 of history[%d] is %d but %d", i, j, d.history.src0[j], src0)
				}
				if d.history.src1[j] != src1 {
					t.Errorf("data[%d]'s expected src1 of history[%d] is %d but %d", i, j, d.history.src1[j], src1)
				}
				if d.history.dst0[j] != dst0 {
					t.Errorf("data[%d]'s expected dst0 of history[%d] is %d but %d", i, j, d.history.dst0[j], dst0)
				}
				if d.history.dst1[j] != dst1 {
					t.Errorf("data[%d]'s expected dst1 of history[%d] is %d but %d", i, j, d.history.dst1[j], dst1)
				}
			}
		}
	}
}
