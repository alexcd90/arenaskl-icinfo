package arenaskl

import "sync/atomic"

type links struct {
	nextOffset uint32
	prevOffset  uint32
}

func (l *links) init(preOffset, nextOffset uint32) {
	l.prevOffset = preOffset
	l.nextOffset = nextOffset
}

type node struct {
	// Immutable fields, so no need to lock to access key.
	keyOffset uint32
	keySize   uint32

	// Multiple parts of the value are encoded as a single uint64 so that
	// it can be atomically loaded and stored
	// value offset: uint32(bits 0-31)
	// value size  : uint16(bits 32-47)
	// metadata    : uint16(bits 48-63)
	value uint64

	// Most nodes do not need to use the full height of the tower, since the
	// probability of each successive level decrease exponentially. Because
	// these elements are never accessed, they do not need to be allocated.
	// Therefore, when a node is allocated in the arena, its memory footprint
	// is deliberately truncated to not include unneeded tower elements.
	//
	// All accesses to elements should use CAS operations, with no need to lock.
	tower [maxHeight]links
}

func newNode(arena *Arena, height uint32) (nd *node, err error) {
	if height < 1 || height > maxHeight {
		panic("height cannot be less than one or greater than the max height")
	}

	// Compute the amout of the tower that that will never be used, since the height
	// is less than maxHeight.
	unusedSize := (maxHeight - int(height)) * linksSize

	nodeOffset, err := arena.Alloc(uint32(MaxNodeSize-unusedSize), uint32(unusedSize), Align8)
	if err != nil {
		return
	}

	nd = (*node)(arena.GetPointer(nodeOffset))
	return
}

func (n *node) getKey(arena *Arena) []byte {
	return arena.GetBytes(n.keyOffset, n.keySize)
}

func (n *node) nextOffset(h int) uint32 {
	return atomic.LoadUint32(&n.tower[h].nextOffset)
}

func (n *node) prevOffset(h int) uint32 {
	return atomic.LoadUint32(&n.tower[h].prevOffset)
}

func (n *node) casNextOffset(h int, old, val uint32) bool {
	return atomic.CompareAndSwapUint32(&n.tower[h].nextOffset, old, val)
}

func (n *node) casPrevOffset(h int, old, val uint32) bool {
	return atomic.CompareAndSwapUint32(&n.tower[h].prevOffset, old, val)
}
