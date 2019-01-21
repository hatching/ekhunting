// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package signatures

import (
	"hatching.io/realtime"
	"hatching.io/realtime/events/onemon"
)

type Base struct {
	Trigger func(signature, description, ioc string)
}

func (b *Base) SetTrigger(trigger func(signature, description, ioc string)) {
	b.Trigger = trigger
}

func (b *Base) Init() {
}

func (b *Base) Process(process *onemon.Process) {
}

func Signatures() []realtime.Process {
	var ret []realtime.Process
	ret = append(ret, &ChildProcess{})
	return ret
}
