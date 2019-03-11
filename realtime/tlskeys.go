// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package realtime

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gopkg.in/mgo.v2/bson"
)

type Monitor struct {
	I    int         `bson:"I,omitempty"`
	Type string      `bson:"type,omitempty"`
	Name string      `bson:"name,omitempty"`
	Args interface{} `bson:"args,omitempty"`
}

func ReadPcapTlsSessions(fname string) (map[string]string, error) {
	handle, err := pcap.OpenOffline(fname)
	if err != nil {
		return nil, fmt.Errorf("error opening pcap: %s", err)
	}

	ret := map[string]string{}
	source := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range source.Packets() {
		if layer := packet.Layer(layers.LayerTypeTCP); layer != nil {
			tcp, _ := layer.(*layers.TCP)

			// TODO Can be removed.
			if tcp.SrcPort != 443 && tcp.DstPort != 443 {
				continue
			}

			var tls layers.TLS
			// TODO If we'd to tcp reassembly, then we can check for errors.
			// As the majority of TLS packets will be split up into multiple
			// packets we almost certainly get decoding issues here.
			tls.DecodeFromBytes(tcp.Payload, gopacket.NilDecodeFeedback)
			for _, handshake := range tls.Handshake {
				for _, sh := range handshake.ServerHello {
					// There's not much use for us without a session_id.
					if len(sh.SessionId) == 0 {
						continue
					}
					server_random := hex.EncodeToString(sh.Random)
					session_id := hex.EncodeToString(sh.SessionId)
					ret[server_random] = session_id
				}
			}
		}
	}
	return ret, nil
}

func readAll(r io.Reader, buf []byte) (err error) {
	remain := len(buf)
	for remain > 0 {
		n, err := r.Read(buf)
		if err != nil {
			return err
		}
		remain -= n
		buf = buf[n:]
	}
	return nil
}

func ReadBsonTlsKeys(fname string) (map[string]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("error opening bson: %s", err)
	}

	m := map[int]string{}
	arg := map[int]map[string]int{}
	ret := map[string]string{}
	for {
		size_ := [4]byte{}
		err = readAll(f, size_[:])
		if err == io.EOF {
			return ret, nil
		} else if err != nil {
			break
		}

		size := binary.LittleEndian.Uint32(size_[:])
		buf := make([]byte, size-4)
		err = readAll(f, buf)
		if err == io.EOF {
			return ret, nil
		} else if err != nil {
			break
		}

		obj := &Monitor{}
		bson.Unmarshal(append(size_[:], buf...), &obj)
		args := obj.Args.([]interface{})

		if obj.Type == "info" {
			m[obj.I] = obj.Name
			arg[obj.I] = map[string]int{}
			for idx, key := range args {
				arg[obj.I][key.(string)] = idx
			}
			continue
		}

		switch m[obj.I] {
		case "PRF":
			if args[2] == "key expansion" {
				server_random := args[arg[obj.I]["server_random"]].(string)
				master_secret := args[arg[obj.I]["master_secret"]].(string)
				ret[server_random] = master_secret
			}
		case "Ssl3GenerateKeyMaterial":
			client_random := args[arg[obj.I]["client_random"]].(string)
			server_random := args[arg[obj.I]["server_random"]].(string)
			master_secret := args[arg[obj.I]["master_secret"]].(string)
			if client_random != "" && server_random != "" {
				ret[server_random] = master_secret
			}
		}
	}
	return ret, err
}
