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
	"github.com/hatching/ekhunting/realtime/tracker"
)

type EventServer struct {
	conn    net.Conn
	mux     sync.Mutex
	rbuf    *bufio.Reader
	cwd     string
	sigs    func() []Process
	tracker *tracker.Trackers
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
			AppId  string `json:"appid,omitempty"`
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
			Image   string `json:"image,omitempty"`

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

func New(cwd string, signatures func() []Process) *EventServer {
	return &EventServer{
		cwd:     cwd,
		sigs:    signatures,
		tracker: tracker.New(),
	}
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
		event.Body.Body.Image = process.Image
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
		go es.HandleOnemon(body)
	case "longtermtask":
		go es.HandleOnemon(body)
	case "dumptls":
		go es.DumpTlsKeys(body)
	}

}

func (es *EventServer) ReadMassURLEvents(f *os.File, dispatcher Dispatch) {
	for {
		msg, err := onemon.NextMessage(f)

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
}

func (es *EventServer) ReadLongtermEvents(f *os.File, dispatcher Dispatch) {

	for {
		msg, err := onemon.NextMessage(f)

		if err == io.EOF {
			break
		}

		switch v := msg.(type) {
		case *onemon.Process:
			dispatcher.TrackProcess(v)
		case *onemon.NetworkFlow:
			dispatcher.NetworkFlow(v)
		}
	}
}

func (es *EventServer) HandleOnemon(body EventBody) {
	onemonpath := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", body.Body.TaskId),
		"logs", "onemon.pb",
	)

	tree := es.tracker.GetProcessTree(body.Body.AppId)
	dispatcher := &Dispatch{}
	dispatcher.Init(es, body.Body.TaskId, es.sigs(), tree)

	tries := 0
	for {
		tries++
		fi, err := os.Stat(onemonpath)
		if os.IsNotExist(err) || fi.Size() < 1024 {
			time.Sleep(time.Second)
			return
		}
		if tries > 60 {
			es.Error(body.Body.TaskId, "Timeout while waiting for .pb file to be created")
			break
		}
		if err != nil {
			log.Fatalln("error", err)
		}
		break
	}

	f, err := es.tracker.GetFile(onemonpath, body.Body.AppId)
	if err != nil {
		log.Fatalln("error", err)
	}

	if body.Body.AppId == "" {
		defer f.Close()
	}

	switch body.Event {
	case "massurltask":
		es.ReadMassURLEvents(f, *dispatcher)
		es.Finished(body.Body.TaskId, "massurltask")
	case "longtermtask":
		es.ReadLongtermEvents(f, *dispatcher)
		es.Finished(body.Body.TaskId, "longtermtask")
	}

	fmt.Println("Done")
}

func (es *EventServer) DumpTlsKeys(body EventBody) {
	pcap := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", body.Body.TaskId), "dump.pcap",
	)
	bson := filepath.Join(
		es.cwd, "storage", "analyses", fmt.Sprintf("%d", body.Body.TaskId),
		"logs", fmt.Sprintf("%d.bson", body.Body.LsassPid),
	)

	for _, path := range []string{pcap, bson} {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			es.Error(body.Body.TaskId,
				fmt.Sprintf("Error dumping TLS keys. File does not exist: %s", path))
			return
		}

	}

	pcap_reader, err := es.tracker.GetPcapReader(pcap, body.Body.AppId)
	if err != nil {
		es.Error(body.Body.TaskId, fmt.Sprintf("Error opening PCAP file: %s", err))
		return
	}

	if body.Body.AppId == "" {
		defer pcap_reader.Closer.Close()
	}

	pcap_keys, err := ReadPcapTlsSessions(pcap_reader.Reader)

	if err != nil {
		if err == io.EOF {
			es.Finished(body.Body.TaskId, "dumptls")
			return

		}
		es.Error(body.Body.TaskId, fmt.Sprintf("Error parsing TLS session from pcap: %s", err))
		return
	}

	bson_keys, err := ReadBsonTlsKeys(bson)
	if err != nil {
		es.Error(body.Body.TaskId, fmt.Sprintf("Error parsing TLS master secrets: %s", err))
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
		es.TlsKeys(body.Body.TaskId, tlskeys)
	}
	es.Finished(body.Body.TaskId, "dumptls")
}
