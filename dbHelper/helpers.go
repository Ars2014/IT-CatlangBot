package dbHelper

import "encoding/binary"

func ItoB(value int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(value))
	return b
}
