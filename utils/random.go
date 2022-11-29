package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"math/rand"
)

func RandomBytes(len int) []byte {
	realLen := len / 4
	data := make([]uint32, realLen)
	for i := 0; i < realLen; i++ {
		data[i] = rand.Uint32()
	}
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, data)
	return bytesBuffer.Bytes()
}

// 尽量为12的倍数
func RandomURLBase64(len int) string {
	return base64.URLEncoding.EncodeToString(RandomBytes(len))
}
