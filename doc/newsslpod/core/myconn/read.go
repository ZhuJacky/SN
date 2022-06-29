// Package myconn provides ...
package myconn

import (
	"fmt"
	"io"
)

// 读取指定字节的数据
func Read(r io.Reader, total int) ([]byte, error) {
	var (
		data   = make([]byte, total)
		length = 0
	)

	for {
		n, err := r.Read(data[length:])
		if n > 0 {
			if length += n; length >= total {
				break
			}
		}

		if err != nil {
			if err == io.EOF {
				err = nil
				break
			} else {
				return nil, err
			}
		}

		if n == 0 {
			fmt.Println("myconn.Read，网络异常，读取 0 字节")
			break
		}
	}

	return data[:length], nil
}
