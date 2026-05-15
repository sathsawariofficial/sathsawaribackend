package socket

func isByteSlice(v interface{}) (bVal []byte, ok bool) {
	bVal, ok = v.([]byte)
	return bVal, ok
}
