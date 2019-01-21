// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package signatures

import (
	"log"

	"hatching.io/realtime/events/onemon"
)

type ChildProcess struct {
	Empty
	processes map[int]int
}

func (cp *ChildProcess) Process(process *onemon.Process) {
	if process.Status == onemon.ProcessStatus_Terminate {
		return
	}

	log.Println("child!", process)
}
