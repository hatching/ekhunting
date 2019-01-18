// Copyright (C) 2018-2019 Hatching B.V.
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

	"hatching.io/triage/events/onemon"
)

type EventServer struct {
	conn net.Conn
	rbuf *bufio.Reader
	cwd  string
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

func (e *EventServer) SetCwd(cwd string) {
	e.cwd = cwd
}

func (e *EventServer) Subscribe(events ...string) error {
	sub := Subscribe{}
	sub.Type = "protocol"
	sub.Action = "subscribe"
	sub.Body.Events = events
	body, err := json.Marshal(&sub)
	if err != nil {
		return err
	}
	e.conn.Write(body)
	e.conn.Write([]byte{'\n'})
	return nil
}

func (e *EventServer) Connect(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("error connecting eventserver", err)
	}

	e.conn = conn
	e.rbuf = bufio.NewReader(e.conn)
	go e.Reader()
}

func (e *EventServer) Reader() {
	for {
		line, err := e.rbuf.ReadBytes('\n')
		if err != nil {
			log.Fatalln("error reading first json blob", err)
		}

		log.Println(">", string(line))

		var event Event
		err = json.Unmarshal(line, &event)
		if err != nil {
			log.Fatalln("error", err)
		}

		log.Println("?", event.Body)
		e.Handle(event.Body.Event, event.Body)
	}
}

func (e *EventServer) Handle(event string, body EventBody) {
	if event == "massurltask" {
		log.Println("task!", body.Body.TaskId)
		go e.onemonReader(body.Body.TaskId)
	}
}

func (e *EventServer) onemonReader(taskid int) error {
	var f io.Reader
	var err error
	filepath := filepath.Join(
		e.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid),
		"logs", "onemon.pb",
	)

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

	for {
		msg, err := onemon.NextMessage(f)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch v := m.(type) {
		case *onemon.Process:
			log.Println("process", v)

		case *onemon.Registry:
			log.Println("registry", v)
		}
	}
}
