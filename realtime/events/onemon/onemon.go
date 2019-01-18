// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package onemon

import (
	"errors"
	"io"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

var (
	ErrShortHeader = errors.New("Stream ended before end-of-header")
	ErrShortData   = errors.New("Stream ended before end-of-data")
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

func NextEvent(r io.Reader) (kind int, data []byte, err error) {
	var header [4]byte
	// 3 byte size
	// 1 byte kind
	// <protobuf>
	n, err := r.Read(header[:])
	if err != nil {
		return
	} else if n != 4 {
		err = ErrShortHeader
		return
	}
	sz := varint(header[:3])
	kind = int(header[3])
	data, err = ioutil.ReadAll(io.LimitReader(r, int64(sz)))
	if err != nil {
		return
	} else if len(data) != sz {
		err = ErrShortData
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
	}
	return nil
}
