// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package signatures

import (
	"regexp"

	"github.com/hatching/ekhunting/realtime/events/onemon"
)

type ChildProcess struct {
	Base
	image map[uint64]string
}

var initial = "C:\\Program Files\\Internet Explorer\\iexplore.exe"
var iexplore = map[string]bool{
	"C:\\Program Files\\Internet Explorer\\iexplore.exe":       true,
	"C:\\Program Files (x86)\\Internet Explorer\\IEXPLORE.EXE": true,
}
var firefox = map[string]bool{
	"C:\\Program Files (x86)\\Mozilla Firefox\\firefox.exe": true,
}
var whitelist_ie = map[string]bool{
	"C:\\Windows\\System32\\ie4uinit.exe": true,
}
var whitelist_ff = map[string]bool{
	"C:\\Program Files (x86)\\Mozilla Firefox\\uninstall\\helper.exe": true,
}

var (
	ieRundll32Process = regexp.MustCompile(
		`^C:\\Windows\\system32\\inetcpl.cpl,ClearMyTracksByProcess Flags:\d+ WinX:0 WinY:0 IEFrame:0000000000000000`,
	)
)

func (sig *ChildProcess) Init() {
	sig.image = map[uint64]string{}
}

func (sig *ChildProcess) Process(process *onemon.Process) {
	if process.Status == onemon.ProcessStatus_Existing {
		return
	}
	if process.Status == onemon.ProcessStatus_Ignore {
		return
	}
	if process.Status == onemon.ProcessStatus_Terminate {
		delete(sig.image, process.Pid)
		return
	}

	sig.image[process.Pid] = process.Image

	// Internet Explorer process.
	if _, ok := iexplore[process.Image]; ok {
		return
	}

	// Firefox process.
	if _, ok := firefox[process.Image]; ok {
		return
	}

	// Not a tracked process (ignore for now).
	if _, ok := sig.image[process.Ppid]; !ok {
		return
	}

	// Whitelisted child process of Internet Explorer.
	if _, ok := iexplore[sig.image[process.Ppid]]; ok {
		if _, ok := whitelist_ie[process.Image]; ok {
			return
		}
		if process.Image == "C:\\Windows\\system32\\rundll32.exe" &&
			ieRundll32Process.MatchString(process.Command) {
			return
		}
	}

	// Whitelisted child process of Firefox.
	if _, ok := firefox[sig.image[process.Ppid]]; ok {
		if _, ok := whitelist_ff[process.Image]; ok {
			return
		}
	}

	sig.Trigger(
		"child_process", "A malicious process was started", process.Command,
	)
}
