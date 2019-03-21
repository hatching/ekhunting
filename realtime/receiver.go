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

	"github.com/hatching/ekhunting/realtime/events/onemon"
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
			Error  string `json:"error,omitempty"`

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

			// DumpTlsKeys event (request).
			LsassPid int `json:"lsass_pid,omitempty"`
			// DumpTlsKeys event (response).
			TlsKeys []TlsKeys `json:"tlskeys,omitempty"`

			// Javascript event (+pid).
			Code string `json:"code,omitempty"`
			Meta string `json:"meta,omitempty"`
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
	TlsKeys struct {
		SessionID    string `json:"session_id,omitempty"`
		MasterSecret string `json:"master_secret,omitempty"`
	}
)

func (es *EventServer) SetSignatures(signatures func() []Process) {
	es.sigs = signatures
}

func (es *EventServer) SetCwd(cwd string) {
	es.cwd = cwd
}

func (es *EventServer) sendEvent(event interface{}) {
	blob, err := json.Marshal(event)
	if err != nil {
		log.Fatalln("error marshalling event", err)
	}

	es.mux.Lock()
	es.conn.Write(blob)
	es.conn.Write([]byte{'\n'})
	es.mux.Unlock()
}

func (es *EventServer) Subscribe(events ...string) {
	event := Subscribe{}
	event.Type = "protocol"
	event.Action = "subscribe"
	event.Body.Events = events
	es.sendEvent(event)
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
	es.sendEvent(event)
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
	es.sendEvent(event)
}

func (es *EventServer) Javascript(taskid int, code, meta string, process *onemon.Process) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "javascript"
	event.Body.Body.TaskId = taskid
	event.Body.Body.Code = code
	event.Body.Body.Meta = meta
	if process != nil {
		event.Body.Body.Pid = int(process.Pid)
		event.Body.Body.Ppid = int(process.Ppid)
	}
	es.sendEvent(event)
}

func (es *EventServer) TlsKeys(taskid int, tlskeys map[string]string) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "tlskeys"
	event.Body.Body.TaskId = taskid
	event.Body.Body.TlsKeys = []TlsKeys{}
	for session_id, master_secret := range tlskeys {
		event.Body.Body.TlsKeys = append(event.Body.Body.TlsKeys, TlsKeys{
			SessionID: session_id, MasterSecret: master_secret,
		})
	}

	es.sendEvent(event)
}

func (es *EventServer) Error(taskid int, err string) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "error"
	event.Body.Body.TaskId = taskid
	event.Body.Body.Error = err
	es.sendEvent(event)
}

func (es *EventServer) Finished(taskid int, action string) {
	// If not running in realtime.
	if es.conn == nil {
		return
	}

	event := Event{}
	event.Type = "event"
	event.Body.Event = "finished"
	event.Body.Body.TaskId = taskid
	event.Body.Body.Action = action
	es.sendEvent(event)
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
	switch body.Event {
	case "massurltask":
		go es.OnemonReaderTask(body.Body.TaskId)
	case "dumptls":
		go es.DumpTlsKeys(body.Body.TaskId, body.Body.LsassPid)
	}
}

func (es *EventServer) OnemonReaderTask(taskid int) {
	filepath := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid),
		"logs", "onemon.pb",
	)
	es.OnemonReaderPath(taskid, filepath)
}

func (es *EventServer) OnemonReaderPath(taskid int, filepath string) {
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
		case *onemon.SyscallS, *onemon.SyscallSS:
			dispatcher.Syscall(v)
		}
	}
	es.Finished(taskid, "massurltask")
}

func (es *EventServer) DumpTlsKeys(taskid, lsasspid int) {
	pcap := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid), "dump.pcap",
	)
	bson := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", taskid),
		"logs", fmt.Sprintf("%d.bson", lsasspid),
	)
	es.DumpTlsKeysPath(taskid, pcap, bson)
}

func (es *EventServer) DumpTlsKeysPath(taskid int, pcap, bson string) {
	pcap_keys, err1 := ReadPcapTlsSessions(pcap)
	bson_keys, err2 := ReadBsonTlsKeys(bson)
	if err1 != nil || err2 != nil {
		es.Error(taskid, fmt.Sprintf("error parsing tls master secrets: %s %s", err1, err2))
		return
	}

	// Session ID -> TLS Master Secret
	tlskeys := map[string]string{}
	if pcap_keys != nil && bson_keys != nil {
		for server_random, session_id := range pcap_keys {
			if master_secret, ok := bson_keys[server_random]; ok {
				tlskeys[session_id] = master_secret
			}
		}
		// TODO Probably not necessary, but iterate both ways just to be sure.
		for server_random, master_secret := range bson_keys {
			if session_id, ok := pcap_keys[server_random]; ok {
				tlskeys[session_id] = master_secret
			}
		}
		es.TlsKeys(taskid, tlskeys)
	}
	es.Finished(taskid, "dumptls")
}
