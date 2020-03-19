package transform

import (
	"bytes"
	"unicode/utf8"

	"golang.org/x/text/transform"
)

// Replacer replaces a part of byte data which matches given pattern to other pattern.
// It implements transform.Transformer.
type Replacer struct {
	old, new []byte
	history  *ReplaceHistory
	preDst   []byte
	preSrc   []byte // preSrc always points subslice of old.
	// offDst and offSrc is the length of transformed bytes until the current Transform call.
	offDst int
	offSrc int
}

var _ transform.Transformer = (*Replacer)(nil)

// NewReplacer creates a new Replacer which replaces old to new.
// old and new are accepted nil and empty bytes ([]byte{}).
// if old is empty the Replacer does not replace and just copy src to dst.
//
// If history is not nil, Replacer records histories of replacing.
func NewReplacer(old, new []byte, history *ReplaceHistory) *Replacer {
	return &Replacer{
		new:     new,
		old:     old,
		history: history,
	}
}

// Reset implements transform.Transformer.Reset.
func (r *Replacer) Reset() {
	r.preDst = nil
	r.preSrc = nil
	r.offDst = 0
	r.offSrc = 0
}

// Transform implements transform.Transformer.Transform.
// Transform replaces old to new in src and copy to dst.
//
// Because the transforming is taken by part of source data with transform.Reader
// the Replacer is carefull for boundary of current src buffer and next one.
// When end of src matches for part of old and atEOF is false
// the Replacer stops to transform and remain the matched bytes for next transforming.
// If Replacer remained boundary bytes, nSrc will be less than len(src)
// and returns transform.ErrShortSrc.
func (r *Replacer) Transform(dst, src []byte, atEOF bool) (int, int, error) {

	_src := src
	if len(r.preSrc) > 0 {
		_src = make([]byte, len(r.preSrc)+len(src))
		copy(_src, r.preSrc)
		copy(_src[len(r.preSrc):], src)
	}

	nDst, nSrc, preSrc, err := r.transform(dst, _src, atEOF)
	r.offDst += nDst
	r.offSrc += nSrc - len(preSrc)

	if nSrc < len(r.preSrc) {
		r.preSrc = r.preSrc[nSrc:]
		nSrc = 0
	} else {
		nSrc -= len(r.preSrc)
		r.preSrc = preSrc
	}

	return nDst, nSrc, err
}

func (r *Replacer) transform(dst, src []byte, atEOF bool) (nDst, nSrc int, preSrc []byte, err error) {
	if len(r.preDst) > 0 {
		n := copy(dst, r.preDst)
		nDst += n
		r.preDst = r.preDst[n:]
		if len(r.preDst) > 0 {
			err = transform.ErrShortDst
			return
		}
	}

	if len(r.old) == 0 {
		n := copy(dst[nDst:], src)
		nDst += n
		nSrc += n
		return
	}

	for {
		i := bytes.Index(src[nSrc:], r.old)

		if i == -1 { // not found
			n := len(src[nSrc:])

			var w int
			if !atEOF {
				w = overlapWidth(src[nSrc:], r.old)
				if w > 0 {
					// exclude w bytes because they may match r.old with next several bytes
					n -= w
					err = transform.ErrShortSrc
				}
			}

			m := copy(dst[nDst:], src[nSrc:nSrc+n])
			nDst += m
			nSrc += m
			if m < n {
				err = transform.ErrShortDst
				return
			}
			preSrc = r.old[:w]
			nSrc += w
			return
		}

		// Copy to i
		n := copy(dst[nDst:], src[nSrc:nSrc+i])
		nDst += n
		nSrc += n
		if n < i {
			err = transform.ErrShortDst
			return
		}

		// Copy new
		r.history.add(r.offSrc+nSrc, r.offSrc+nSrc+len(r.old), r.offDst+nDst, r.offDst+nDst+len(r.new))
		n = copy(dst[nDst:], r.new)
		nDst += n
		nSrc += len(r.old)
		if n < len(r.new) {
			r.preDst = r.new[n:]
			err = transform.ErrShortDst
			return
		}
	}
}

// overlapWidth returns the length of longest match of end of a and start of b.
// Returns 0 if no match.
func overlapWidth(a, b []byte) int {
	w := len(a)
	if w > len(b) {
		w = len(b)
	}
	for ; w > 0; w-- {
		if bytes.Equal(a[len(a)-w:], b[:w]) {
			return w
		}
	}
	return 0
}

// Replace returns a Replacer with out history.
// It is a shorthand for NewReplacer(old, new, nil).
func Replace(old, new []byte) *Replacer {
	return NewReplacer(old, new, nil)
}

// ReplaceRune returns a Replacer which replaces given rune.
func ReplaceRune(old, new rune) *Replacer {
	oldBuf := make([]byte, utf8.RuneLen(old))
	utf8.EncodeRune(oldBuf, old)

	newBuf := make([]byte, utf8.RuneLen(new))
	utf8.EncodeRune(newBuf, new)

	return Replace(oldBuf, newBuf)
}

// ReplaceString returns a Replacer which replaces given string.
func ReplaceString(old, new string) *Replacer {
	return Replace([]byte(old), []byte(new))
}

// ReplaceTable is used for ReplaceAll.
type ReplaceTable interface {
	// At returns i-th replacing rule.
	At(i int) (old, new []byte)
	// Len returns the number of replacing rules.
	Len() int
}

// ReplaceByteTable implements ReplaceTable.
// i*2 elements represents old, i*2+1 elements new for Replacer.
type ReplaceByteTable [][]byte

// Add adds a new replacing rule.
func (t *ReplaceByteTable) Add(old, new []byte) {
	*t = append(*t, old, new)
}

// At implements ReplaceTable.At.
func (t ReplaceByteTable) At(i int) (old, new []byte) {
	return t[i*2], t[i*2+1]
}

// Len implements ReplaceTable.Len.
func (t ReplaceByteTable) Len() int {
	return len(t) / 2
}

// ReplaceStringTable implements ReplaceTable.
// i*2 elements represents old, i*2+1 elements new for Replacer.
type ReplaceStringTable []string

// Add adds a new replacing rule.
func (t *ReplaceStringTable) Add(old, new string) {
	*t = append(*t, old, new)
}

// At implements ReplaceTable.At.
func (t ReplaceStringTable) At(i int) (old, new []byte) {
	return []byte(t[i*2]), []byte(t[i*2+1])
}

// Len implements ReplaceTable.Len.
func (t ReplaceStringTable) Len() int {
	return len(t) / 2
}

// ReplaceRuneTable implements ReplaceTable.
// i*2 elements represents old, i*2+1 elements new for Replacer.
type ReplaceRuneTable []rune

// Add adds a new replacing rule.
func (t *ReplaceRuneTable) Add(old, new rune) {
	*t = append(*t, old, new)
}

// At implements ReplaceTable.At.
func (t ReplaceRuneTable) At(i int) (old, new []byte) {
	old = make([]byte, utf8.RuneLen(t[i*2]))
	utf8.EncodeRune(old, t[i*2])

	new = make([]byte, utf8.RuneLen(t[i*2+1]))
	utf8.EncodeRune(new, t[i*2+1])

	return old, new
}

// Len implements ReplaceTable.Len.
func (t ReplaceRuneTable) Len() int {
	return len(t) / 2
}

// ReplaceAll creates transform.Transformer which is chained Replacers.
// The Replacers replace by replacing rule which is indicated by ReplaceTable.
func ReplaceAll(t ReplaceTable) transform.Transformer {
	rs := make([]transform.Transformer, t.Len())
	for i := range rs {
		old, new := t.At(i)
		rs[i] = Replace(old, new)
	}
	return transform.Chain(rs...)
}
