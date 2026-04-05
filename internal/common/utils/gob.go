package utils

import (
	"bytes"
	"encoding/gob"
)

// GobEncoder Gob 通用编码工具，将任意类型编码为字节切片
//
// 参数:
//   - v interface{} 待编码的数据
//
// 返回值:
//   - []byte 编码后的字节切片
//   - error 编码失败时返回错误
func GobEncoder(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}

// GobDecoder Gob 通用解码工具，将字节切片解码到目标变量
//
// 参数:
//   - b []byte 待解码的字节切片
//   - v interface{} 解码目标，需为指针类型
//
// 返回值:
//   - error 解码失败时返回错误
func GobDecoder(b []byte, v interface{}) error {
	decoder := gob.NewDecoder(bytes.NewReader(b))
	return decoder.Decode(v)
}
