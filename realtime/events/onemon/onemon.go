// Copyright (C) 2018-2019 Hatching B.V.
// All rights reserved.

package onemon

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/golang/protobuf/proto"
)

var (
	ErrUnsupported = errors.New("Unsupported message type")
)

func NextMessage(f *os.File) (interface{}, error) {
	kind, data, err := NextEvent(f)
	if err != nil {
		return nil, err
	}
	e := MessageByType(kind)
	if e == nil {
		return nil, ErrUnsupported
	}
	err = proto.Unmarshal(data, e)
	if err != nil {
		fmt.Println("Error during unmarshal:", err)
	}
	return e, err
}

func readAll(f *os.File, buf []byte) (read int, err error) {
	remain := len(buf)
	for remain > 0 {
		n, err := f.Read(buf)
		read += n
		if err != nil {
			return read, err
		}
		remain -= n
		buf = buf[n:]
	}
	return read, nil
}

func SetPreviousPos(f *os.File, hsz, bsz int) {
	f.Seek(-int64(hsz+bsz), io.SeekCurrent)
}

func NextEvent(f *os.File) (kind int, data []byte, err error) {
	header := make([]byte, 4)
	// 3 byte size
	// 1 byte kind
	// <protobuf>
	read, err := readAll(f, header)
	if err != nil {
		if err == io.EOF {
			SetPreviousPos(f, read, 0)
		}
		return 0, nil, err
	}
	sz := varint(header[:3])

	kind = int(header[3])
	data = make([]byte, sz)
	read, err = readAll(f, data)
	if err == io.EOF || read < sz {
		SetPreviousPos(f, len(header), read)
	}
	if err != nil {
		return
	}
	return
}

func varint(b []byte) (r int) {
	var m int = 1
	for _, v := range b {
		r += int(v) * m
		m *= 256
	}
	return
}

func MessageByType(kind int) proto.Message {
	switch kind {
	case 1:
		return &Process{}
	case 2:
		return &Registry{}
	case 8:
		return &File{}
	case 12:
		return &NetworkFlow{}
	case 102:
		return &SyscallS{}
	case 103:
		return &SyscallSS{}
	}
	return nil
}
