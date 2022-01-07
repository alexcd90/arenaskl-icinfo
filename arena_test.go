package arenaskl

import (
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestArenaSizeOverflow(t *testing.T)  {
	a := NewArena(math.MaxUint32)

	// Allocating under the limit throws no error
	offset, err := a.Alloc(math.MaxUint16, 0, Align1)
	require.Nil(t, err)
	require.Equal(t, uint32(1), offset)
	require.Equal(t, uint32(math.MaxUint16)+1, a.Size())

	//Allocating over the limit could cause an accounting
	//overflow if 32-bit arithmetic was used. It should't.
	_, err = a.Alloc(math.MaxUint32, 0, Align1)
	require.Equal(t, ErrArenaFull, err)
	require.Equal(t, uint32(math.MaxUint32), a.Size())

	// Continuing to allocate continues to throw an error.
	_, err = a.Alloc(math.MaxUint16, 0, Align1)
	require.Equal(t, ErrArenaFull, err)
	require.Equal(t, uint32(math.MaxUint32), a.Size())

}
