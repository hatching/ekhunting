package tracker

import (
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hatching/ekhunting/realtime/events/onemon"
	"github.com/hatching/gopacket/pcapgo"
)

type (
	PcapReader struct {
		Closer io.Closer
		Reader *pcapgo.Reader
	}

	Tracker struct {
		paths      map[string]*os.File
		pcapreader *PcapReader
		tree       map[int]*onemon.Process
		used       time.Time
	}

	Trackers struct {
		apps   map[string]*Tracker
		appmux sync.Mutex
	}
)

func New() *Trackers {
	t := &Trackers{
		apps: make(map[string]*Tracker),
	}
	go t.cleanFiles()
	return t
}

func (t *Tracker) close() {
	if t.pcapreader != nil {
		err := t.pcapreader.Closer.Close()
		if err != nil {
			log.Println("Error closing PCAP reader", err)
		}
		for k, f := range t.paths {
			err := f.Close()
			if err != nil {
				log.Println("Error closing file", k, err)
			}
		}
	}

}

func (t *Trackers) cleanFiles() {
	for {
		time.Sleep(time.Second * 30)
		t.appmux.Lock()

		now := time.Now()
		for k, tracker := range t.apps {
			if diff := now.Sub(tracker.used); diff > time.Minute*15 {
				tracker.close()
				log.Println("Cleaning tracker for appid:", k)
				delete(t.apps, k)
			}
		}
		t.appmux.Unlock()
	}
}

// Return a *os.File for a given file path. If an app ID is provided, a previously
// used reader will be returned for that given path and ID.
func (t *Trackers) GetFile(filepath, appid string) (*os.File, error) {
	if appid == "" {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	t.appmux.Lock()
	defer t.appmux.Unlock()

	if _, ok := t.apps[appid]; !ok {
		t.CreateTracker(appid)
	}

	tracker := t.apps[appid]

	if _, ok := tracker.paths[filepath]; !ok {
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}

		tracker.paths[filepath] = f
	}

	tracker.used = time.Now()
	return tracker.paths[filepath], nil
}

func (t *Trackers) GetPcapReader(pcap_path, appid string) (*PcapReader, error) {
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
			Closer: f,
			Reader: reader,
		}, nil
	}

	t.appmux.Lock()
	defer t.appmux.Unlock()

	if _, ok := t.apps[appid]; !ok {
		t.CreateTracker(appid)
	}
	tracker := t.apps[appid]

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
			Closer: f,
			Reader: reader,
		}
	}
	tracker.used = time.Now()

	return tracker.pcapreader, nil
}

func (t *Trackers) GetProcessTree(appid string) map[int]*onemon.Process {
	if appid == "" {
		return map[int]*onemon.Process{}
	}

	t.appmux.Lock()
	defer t.appmux.Unlock()

	if _, ok := t.apps[appid]; !ok {
		t.CreateTracker(appid)
	}
	tracker := t.apps[appid]

	if tracker.tree == nil {
		tracker.tree = map[int]*onemon.Process{}
	}
	tracker.used = time.Now()

	return tracker.tree
}

func (t *Trackers) CreateTracker(appid string) {
	t.apps[appid] = &Tracker{
		paths: make(map[string]*os.File),
		used:  time.Now(),
	}
}
