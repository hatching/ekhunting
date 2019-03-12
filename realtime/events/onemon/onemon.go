// Copyright (C) 2018-2019 Hatching B.V.
// All rights reserved.

package onemon

import (
	"errors"
	"io"

	"github.com/golang/protobuf/proto"
)

var (
	ErrUnsupported = errors.New("Unsupported message type")
)

func NextMessage(r io.Reader) (interface{}, error) {
	kind, data, err := NextEvent(r)
	if err != nil {
		return nil, err
	}
	e := MessageByType(kind)
	if e == nil {
		return nil, ErrUnsupported
	}
	err = proto.Unmarshal(data, e)
	return e, err
}

func readAll(r io.Reader, buf []byte) (err error) {
	remain := len(buf)
	for remain > 0 {
		n, err := r.Read(buf)
		if err != nil {
			return err
		}
		remain -= n
		buf = buf[n:]
	}
	return nil
}

func NextEvent(r io.Reader) (kind int, data []byte, err error) {
	var header [4]byte
	// 3 byte size
	// 1 byte kind
	// <protobuf>
	err = readAll(r, header[:])
	if err != nil {
		return 0, nil, err
	}
	sz := varint(header[:3])
	kind = int(header[3])
	data = make([]byte, sz)
	err = readAll(r, data)
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
	case 12:
		return &NetworkFlow{}
	case 102:
		return &SyscallS{}
	case 103:
		return &SyscallSS{}
	}
	return nil
}
