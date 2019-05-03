package transform_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"golang.org/x/text/transform"

	. "github.com/tenntenn/text/transform"
)

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
			nSrc:     10,
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
			nSrc:     9,
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

// TestReplacerWithReader is a test for transform.Replace with transform.Reader.
func TestReplacerWithReader(t *testing.T) {
	type history struct {
		src0, src1 []int
		dst0, dst1 []int
	}

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
