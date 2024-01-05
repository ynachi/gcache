package frame

import (
	"fmt"
	"io"
	"strings"
)

// Array represents an array of Frames, which can be mixed types.
// It looks redundant to set the size as it could be deduced from the length
// of the slice this is needed. For example, to decode an Array from a buffer, we
// will need to know its length in advance.
type Array struct {
	size  int
	value []Framer
}

// NewArray create a new Array. This array is not reallocated (means its length is 0) but the
// capacity is set to avoid sliding. The reason the initial len is 0 is that we expect it to be
// filled via append method.
func NewArray(size int) *Array {
	value := make([]Framer, 0, size)
	return &Array{size: size, value: value}
}

// Append add a new frame to an Array. It gerror when there is not enough capacity to add more.
// It does not grow the Array automatically.
func (a *Array) Append(f Framer) error {
	if len(a.value) >= a.size {
		return ErrArrayIsFull
	}
	a.value = append(a.value, f)
	return nil
}

// String provides a text representation of an  Array frame.
func (a *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(a.value)))
	for _, f := range a.value {
		sb.WriteString(f.String())
	}
	return sb.String()
}

func (a *Array) Size() int {
	return a.size
}

// Get return the Frame at position i in the Array. It would panic if the index is out of bounds.
func (a *Array) Get(i int) Framer {
	return a.value[i]
}

func (a *Array) Serialize() []byte {
	return []byte(a.String())
}

func (a *Array) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := a.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}
