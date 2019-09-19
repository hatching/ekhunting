// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
	"encoding/base64"
	"net"

	"github.com/hatching/ekhunting/realtime/events/onemon"
)

type Process interface {
	SetTrigger(func(signature, description, ioc string))
	Init()
	Process(process *onemon.Process)
}

type Dispatch struct {
	es         *EventServer
	taskid     int
	signatures []Process
	tree       map[int]*onemon.Process
}

func (d *Dispatch) Init(es *EventServer, taskid int, signatures []Process, tree map[int]*onemon.Process) {
	d.taskid = taskid
	d.es = es
	d.signatures = signatures
	d.tree = tree
	for _, signature := range d.signatures {
		signature.SetTrigger(d.Trigger)
		signature.Init()
	}
}

func (d *Dispatch) TrackProcess(process *onemon.Process) {
	d.tree[int(process.Pid)] = process
	d.es.Process(
		int(process.Ts), d.taskid, int(process.Pid), int(process.Ppid),
		process.Status.String(), process.Image, process.Command, process.Orig,
	)
}

func (d *Dispatch) Process(process *onemon.Process) {
	d.tree[int(process.Pid)] = process
	for _, signature := range d.signatures {
		signature.Process(process)
	}
}

func (d *Dispatch) File(file *onemon.File) {
	d.es.File(
		int(file.Ts), d.taskid, int(file.Pid), file.Kind.String(), file.Srcpath,
		file.Dstpath,
	)
}

func (d *Dispatch) Registry(reg *onemon.Registry) {
	values := reg.Values
	op := reg.Kind.String()
	if op == "SetValueKeyDat" {
		values = base64.StdEncoding.EncodeToString(reg.Valued)
	}
	d.es.registry(
		int(reg.Ts), d.taskid, int(reg.Pid), int(reg.Valuei), op, reg.Path,
		values,
	)
}

func int2ipv4(ip uint32) net.IP {
	return net.IPv4(uint8(ip), uint8(ip>>8), uint8(ip>>16), uint8(ip>>24))
}

func (d *Dispatch) NetworkFlow(netflow *onemon.NetworkFlow) {
	d.es.NetworkFlow(
		int(netflow.Ts), d.taskid, int(netflow.Proto), int2ipv4(netflow.Srcip),
		int2ipv4(netflow.Dstip), int(netflow.Srcport), int(netflow.Dstport),
		d.tree[int(netflow.Pid)],
	)
}

func (d *Dispatch) Syscall(obj interface{}) {
	switch v := obj.(type) {
	case *onemon.SyscallS:
		if v.Kind == onemon.SyscallSKind_JsGlobalObjectDefaultEvalHelper {
			d.es.Javascript(d.taskid, v.Arg0, "no context", d.tree[int(v.Pid)])
		}
	case *onemon.SyscallSS:
		if v.Kind == onemon.SyscallSSKind_COleScript_Compile {
			d.es.Javascript(d.taskid, v.Arg0, v.Arg1, d.tree[int(v.Pid)])
		}
	}
}

func (d *Dispatch) Trigger(signature, description, ioc string) {
	d.es.Trigger(d.taskid, signature, description, ioc)
}
