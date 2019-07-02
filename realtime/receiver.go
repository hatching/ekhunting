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
	"github.com/hatching/gopacket/pcapgo"
)

type PcapReader struct {
	closer io.Closer
	r      *pcapgo.Reader
}

type FileReader struct {
	closer io.Closer
	r      *bufio.Reader
}

type Tracker struct {
	paths      map[string]*FileReader
	pcapreader *PcapReader
	used       time.Time
}

type EventServer struct {
	conn   net.Conn
	mux    sync.Mutex
	rbuf   *bufio.Reader
	cwd    string
	sigs   func() []Process
	apps   map[string]*Tracker
	appmux sync.Mutex
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

func (es *EventServer) InitApps() {
	es.apps = make(map[string]*Tracker)
	go es.cleanFiles()
}

func (es *EventServer) SetSignatures(signatures func() []Process) {
	es.sigs = signatures
}

func (es *EventServer) SetCwd(cwd string) {
	es.cwd = cwd
}

func (es *EventServer) cleanFiles() {
	for {
		time.Sleep(time.Second * 30)
		es.appmux.Lock()

		now := time.Now()
		for k, tracker := range es.apps {
			if diff := now.Sub(tracker.used); diff > time.Minute*15 {
				tracker.close()
				fmt.Println("Cleaning tracker for appid:", k)
				delete(es.apps, k)
			}
		}
		es.appmux.Unlock()
	}
}

func (t *Tracker) close() {
	if t.pcapreader != nil {
		err := t.pcapreader.closer.Close()
		if err != nil {
			log.Println("Error closing PCAP reader", err)
		}
		for k, fr := range t.paths {
			err := fr.closer.Close()
			if err != nil {
				log.Println("Error closing file", k, err)
			}
		}
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

// Return a FileReader for a given file path. If an app ID is provided, a previously
// used reader will be returned for that given path and ID.
func (es *EventServer) GetFileReader(filepath, appid string) (*FileReader, error) {
	if appid == "" {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		return &FileReader{
			closer: f,
			r:      bufio.NewReader(f),
		}, nil
	}

	es.appmux.Lock()
	defer es.appmux.Unlock()

	if _, ok := es.apps[appid]; !ok {
		es.CreateTracker(appid)
	}

	tracker := es.apps[appid]

	if _, ok := tracker.paths[filepath]; !ok {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}

		tracker.paths[filepath] = &FileReader{
			closer: f,
			r:      bufio.NewReader(f),
		}
	}

	tracker.used = time.Now()
	return tracker.paths[filepath], nil
}

func (es *EventServer) GetPcapReader(pcap_path, appid string) (*PcapReader, error) {
	if appid == "" {
		f, err := os.Open(pcap_path)
		if err != nil {
			return nil, err
		}
		reader, err := pcapgo.NewReader(f)
		if err != nil {
			return nil, err
		}
		return &PcapReader{
			closer: f,
			r:      reader,
		}, nil
	}

	es.appmux.Lock()
	defer es.appmux.Unlock()

	if _, ok := es.apps[appid]; !ok {
		es.CreateTracker(appid)
	}
	tracker := es.apps[appid]

	if tracker.pcapreader == nil {
		f, err := os.Open(pcap_path)
		if err != nil {
			return nil, err
		}
		reader, err := pcapgo.NewReader(f)
		if err != nil {
			return nil, err
		}
		tracker.pcapreader = &PcapReader{
			closer: f,
			r:      reader,
		}
	}
	tracker.used = time.Now()

	return tracker.pcapreader, nil
}

func (es *EventServer) CreateTracker(appid string) {
	es.apps[appid] = &Tracker{
		paths: make(map[string]*FileReader),
		used:  time.Now(),
	}
}

func (es *EventServer) ReadMassURLEvents(r *bufio.Reader, dispatcher Dispatch) {
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
}

func (es *EventServer) ReadLongtermEvents(r *bufio.Reader, dispatcher Dispatch) {
	for {
		msg, err := onemon.NextMessage(r)

		if err == io.EOF {
			break
		}

		switch v := msg.(type) {
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

	dispatcher := &Dispatch{}
	dispatcher.Init(es, body.Body.TaskId, es.sigs())

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
			return
		}
		if err != nil {
			log.Fatalln("error", err)
		}
		break
	}

	fr, err := es.GetFileReader(onemonpath, body.Body.AppId)
	if err != nil {
		log.Fatalln("error", err)
	}

	if body.Body.AppId == "" {
		defer fr.closer.Close()
	}

	if b, _ := fr.r.Peek(4); string(b) == "FILE" {
		// Skip file header
		fr.r.ReadLine()
		fr.r.ReadLine()
	}

	switch body.Event {
	case "massurltask":
		es.ReadMassURLEvents(fr.r, *dispatcher)
		es.Finished(body.Body.TaskId, "massurltask")
	case "longtermtask":
		es.ReadLongtermEvents(fr.r, *dispatcher)
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

	pcap_reader, err := es.GetPcapReader(pcap, body.Body.AppId)
	if err != nil {
		es.Error(body.Body.TaskId, fmt.Sprintf("Error opening PCAP file: %s", err))
		return
	}

	if body.Body.AppId == "" {
		defer pcap_reader.closer.Close()
	}

	pcap_keys, err := ReadPcapTlsSessions(pcap_reader.r)
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
