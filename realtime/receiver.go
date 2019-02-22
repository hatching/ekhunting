// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hatching.io/realtime/events/onemon"
)

type EventServer struct {
	conn net.Conn
	mux  sync.Mutex
	rbuf *bufio.Reader
	cwd  string
	sigs func() []Process
}

type (
	Header struct {
		Type   string `json:"type"`
		Action string `json:"action,omitempty"`
	}
	EventBody struct {
		Event string `json:"event"`
		Body  struct {
			TaskId int    `json:"taskid"`
			Status string `json:"status,omitempty"`
			Action string `json:"action,omitempty"`

			// Signature event.
			Signature   string `json:"signature,omitempty"`
			Description string `json:"description,omitempty"`
			Ioc         string `json:"ioc,omitempty"`

			// Netflow event.
			Proto   int    `json:"proto,omitempty"`
			Srcip   string `json:"srcip,omitempty"`
			Dstip   string `json:"dstip,omitempty"`
			Srcport int    `json:"srcport,omitempty"`
			Dstport int    `json:"dstport,omitempty"`
			Pid     int    `json:"pid,omitempty"`
			Ppid    int    `json:"ppid,omitempty"`
		} `json:"body"`
	}
	Event struct {
		Header
		Body EventBody `json:"body"`
	}
	Subscribe struct {
		Header
		Body struct {
			Events []string `json:"events"`
		} `json:"body"`
	}
)

func (es *EventServer) SetSignatures(signatures func() []Process) {
	es.sigs = signatures
}

func (es *EventServer) SetCwd(cwd string) {
	es.cwd = cwd
}

func (es *EventServer) Subscribe(events ...string) error {
	sub := Subscribe{}
	sub.Type = "protocol"
	sub.Action = "subscribe"
	sub.Body.Events = events

	body, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	es.mux.Lock()
	es.conn.Write(body)
	es.conn.Write([]byte{'\n'})
	es.mux.Unlock()
	return nil
}

func (es *EventServer) Connect(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("error connecting eventserver", err)
	}

	es.conn = conn
	es.rbuf = bufio.NewReader(es.conn)
	go es.Reader()
}

func (es *EventServer) Trigger(taskid int, signature, description, ioc string) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "signature"
	event.Body.Body.TaskId = taskid
	event.Body.Body.Signature = signature
	event.Body.Body.Description = description
	event.Body.Body.Ioc = ioc

	blob, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("marshal error", err)
	}

	es.mux.Lock()
	es.conn.Write(blob)
	es.conn.Write([]byte{'\n'})
	es.mux.Unlock()
}

func (es *EventServer) NetworkFlow(taskid int, proto int, srcip, dstip net.IP, srcport, dstport int, process *onemon.Process) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "netflow"
	event.Body.Body.TaskId = taskid
	event.Body.Body.Proto = proto
	event.Body.Body.Srcip = srcip.String()
	event.Body.Body.Dstip = dstip.String()
	event.Body.Body.Srcport = srcport
	event.Body.Body.Dstport = dstport
	if process != nil {
		event.Body.Body.Pid = int(process.Pid)
		event.Body.Body.Ppid = int(process.Ppid)
	}

	blob, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("marshal error", err)
	}

	es.mux.Lock()
	es.conn.Write(blob)
	es.conn.Write([]byte{'\n'})
	es.mux.Unlock()
}

func (es *EventServer) Finished(taskid int) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "finished"
	event.Body.Body.TaskId = taskid

	blob, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("marshal error", err)
	}

	es.mux.Lock()
	es.conn.Write(blob)
	es.conn.Write([]byte{'\n'})
	es.mux.Unlock()
}

func (es *EventServer) Reader() {
	for {
		line, err := es.rbuf.ReadBytes('\n')
		if err != nil {
			log.Fatalln("error reading json blob", err)
		}

		log.Println(">", string(line))

		var event Event
		err = json.Unmarshal(line, &event)
		if err != nil {
			log.Fatalln("error parsing json blob", err)
		}

		log.Println("?", event.Body)
		es.Handle(event.Body)
	}
}

func (es *EventServer) Handle(body EventBody) {
	if body.Event == "massurltask" {
		log.Println("task!", body.Body.TaskId)
		go es.OnemonReaderTask(body.Body.TaskId)
	}
}

func (es *EventServer) OnemonReaderTask(taskid int) error {
	filepath := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid),
		"logs", "onemon.pb",
	)
	return es.OnemonReaderPath(taskid, filepath)
}

func (es *EventServer) OnemonReaderPath(taskid int, filepath string) error {
	dispatcher := &Dispatch{}
	dispatcher.Init(es, taskid, es.sigs())

	for {
		fi, err := os.Stat(filepath)
		if os.IsNotExist(err) || fi.Size() < 1024 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Fatalln("error", err)
		}
		break
	}

	f, err := os.Open(filepath)
	if err != nil {
		log.Fatalln("error", err)
	}

	r := bufio.NewReader(f)
	if b, _ := r.Peek(4); string(b) == "FILE" {
		// Skip file header
		r.ReadLine()
		r.ReadLine()
	}

	for {
		msg, err := onemon.NextMessage(r)
		if err == io.EOF {
			break
		}

		switch v := msg.(type) {
		case *onemon.Process:
			dispatcher.Process(v)
		case *onemon.NetworkFlow:
			dispatcher.NetworkFlow(v)
		}
	}

	es.Finished(taskid)
	return nil
}
