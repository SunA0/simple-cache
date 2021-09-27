package cache

type ByteView struct {
	b []byte
}

// ByteSlice
// return CloneBytes(v.B)
func (v ByteView) ByteSlice() []byte {
	return CloneBytes(v.b)
}

// CloneBytes
// copy b
func CloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func (v ByteView) Len() int {
	return len(v.b)
}
