// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package signatures

import (
	"hatching.io/realtime"
	"hatching.io/realtime/events/onemon"
)

type Empty struct {
}

func (e *Empty) Init() {
}

func (e *Empty) Process(process *onemon.Process) {
}

func Signatures() []realtime.Process {
	var ret []realtime.Process
	ret = append(ret, &ChildProcess{})
	return ret
}
