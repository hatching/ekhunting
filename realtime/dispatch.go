// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
	"hatching.io/realtime/events/onemon"
)

type Process interface {
	Init()
	Process(process *onemon.Process)
}

type Dispatch struct {
	signatures []Process
}

func (d *Dispatch) Init(signatures []Process) {
	d.signatures = signatures
	for _, signature := range d.signatures {
		signature.Init()
	}
}

func (d *Dispatch) Process(process *onemon.Process) {
	for _, signature := range d.signatures {
		signature.Process(process)
	}
}
