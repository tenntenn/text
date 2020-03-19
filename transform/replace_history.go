package transform

// ReplaceHistory represents histories of replacing with Replacer.
type ReplaceHistory struct {
	src0, src1 []int
	dst0, dst1 []int
}

// NewReplaceHistory creates a new ReplaceHistory.
func NewReplaceHistory() *ReplaceHistory {
	return &ReplaceHistory{}
}

func (h *ReplaceHistory) add(src0, src1, dst0, dst1 int) {
	// ignore receiver is nil
	if h == nil {
		return
	}

	h.src0 = append(h.src0, src0)
	h.src1 = append(h.src1, src1)
	h.dst0 = append(h.dst0, dst0)
	h.dst1 = append(h.dst1, dst1)
}

// Iterate iterates histories by replacing order.
// This method can call with a nil receiver.
// The arguments of f represent range of replacing, from src[src0:src1] to dst[dst0:dst1].
// if f returns false Iterate will stop the iteration.
func (h *ReplaceHistory) Iterate(f func(src0, src1, dst0, dst1 int) bool) {
	// ignore receiver is nil
	if h == nil {
		return
	}

	for i := range h.src0 {
		if !f(h.src0[i], h.src1[i], h.dst0[i], h.dst1[i]) {
			break
		}
	}
}

// At returns a history of given index.
func (h *ReplaceHistory) At(index int) (src0, src1, dst0, dst1 int) {
	return h.src0[index], h.src1[index], h.dst0[index], h.dst1[index]
}


