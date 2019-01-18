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
		es.Handle(event.Body.Event, event.Body)
	}
}

func (es *EventServer) Handle(event string, body EventBody) {
	if event == "massurltask" {
		log.Println("task!", body.Body.TaskId)
		go es.onemonReader(body.Body.TaskId)
	}
}

func (es *EventServer) onemonReader(taskid int) error {
	var f io.Reader
	var err error

	filepath := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid),
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

		switch msg.(type) {
		case *onemon.Process:
			log.Println("process", msg)
		}
	}
}
