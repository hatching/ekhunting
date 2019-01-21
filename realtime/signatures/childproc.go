// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package signatures

import (
	"log"

	"hatching.io/realtime/events/onemon"
)

type ChildProcess struct {
	Base
	image map[uint64]string
}

var initial = "\\??\\C:\\Program Files\\Internet Explorer\\iexplore.exe"
var iexplore = map[string]bool{
	"\\??\\C:\\Program Files\\Internet Explorer\\iexplore.exe":       true,
	"\\??\\C:\\Program Files (x86)\\Internet Explorer\\IEXPLORE.EXE": true,
}
var whitelist = map[string]bool{
	"\\??\\C:\\Windows\\System32\\ie4uinit.exe": true,
}

func (cp *ChildProcess) Init() {
	cp.image = map[uint64]string{}
}

func (cp *ChildProcess) Process(process *onemon.Process) {
	if process.Status == onemon.ProcessStatus_Terminate {
		return
	}

	cp.image[process.Pid] = process.Image

	// Internet Explorer process.
	if _, ok := iexplore[process.Image]; ok {
		return
	}

	// Not a tracked process (ignore for now).
	if _, ok := cp.image[process.Ppid]; !ok {
		return
	}

	// Whitelisted child process of Internet Explorer.
	if _, ok := iexplore[cp.image[process.Ppid]]; ok {
		if _, ok := whitelist[process.Image]; ok {
			return
		}
	}

	cp.Trigger("child_process", "A malicious process was started", process.Command)
}
