// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
	"net"

	"hatching.io/realtime/events/onemon"
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

func (d *Dispatch) Init(es *EventServer, taskid int, signatures []Process) {
	d.taskid = taskid
	d.es = es
	d.signatures = signatures
	d.tree = map[int]*onemon.Process{}
	for _, signature := range d.signatures {
		signature.SetTrigger(d.Trigger)
		signature.Init()
	}
}

func (d *Dispatch) Process(process *onemon.Process) {
	d.tree[int(process.Pid)] = process
	for _, signature := range d.signatures {
		signature.Process(process)
	}
}

func int2ipv4(ip uint32) net.IP {
	return net.IPv4(uint8(ip), uint8(ip>>8), uint8(ip>>16), uint8(ip>>24))
}

func (d *Dispatch) NetworkFlow(netflow *onemon.NetworkFlow) {
	d.es.NetworkFlow(
		d.taskid, int(netflow.Proto), int2ipv4(netflow.Srcip),
		int2ipv4(netflow.Dstip), int(netflow.Srcport), int(netflow.Dstport),
		d.tree[int(netflow.Pid)],
	)
}

func (d *Dispatch) Trigger(signature, description, ioc string) {
	d.es.Trigger(d.taskid, signature, description, ioc)
}
