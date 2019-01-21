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
	"time"

	"hatching.io/realtime/events/onemon"
)

type EventServer struct {
	conn net.Conn
	rbuf *bufio.Reader
	cwd  string
	sigs func() []Process
}

type (
	Header struct {
		Type   string `json:"type"`
		Action string `json:"action",omitempty`
	}
	EventBody struct {
		Event string `json:"event"`
		Body  struct {
			TaskId int    `json:"taskid"`
			Status string `json:"status"`
			Action string `json:"action"`
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
	body, err := json.Marshal(&sub)
	if err != nil {
		return err
	}

	es.conn.Write(body)
	es.conn.Write([]byte{'\n'})
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
	return es.OnemonReaderPath(filepath)
}

func (es *EventServer) OnemonReaderPath(filepath string) error {
	var f io.Reader
	var err error

	dispatcher := &Dispatch{}
	dispatcher.Init(es.sigs())

	for {
		f, err = os.Open(filepath)
		if os.IsNotExist(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err != nil {
			log.Fatalln("error", err)
		}
		break
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
			return nil
		}

		switch v := msg.(type) {
		case *onemon.Process:
			dispatcher.Process(v)
		}
	}
}
