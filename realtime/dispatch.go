// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
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
}

func (d *Dispatch) Init(es *EventServer, taskid int, signatures []Process) {
	d.taskid = taskid
	d.es = es
	d.signatures = signatures
	for _, signature := range d.signatures {
		signature.SetTrigger(d.Trigger)
		signature.Init()
	}
}

func (d *Dispatch) Process(process *onemon.Process) {
	for _, signature := range d.signatures {
		signature.Process(process)
	}
}

func (d *Dispatch) Trigger(signature, description, ioc string) {
	d.es.Trigger(d.taskid, signature, description, ioc)
}
