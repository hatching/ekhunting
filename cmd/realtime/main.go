// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package main

import (
	"log"
	"os"
	"time"

	"hatching.io/realtime"
	"hatching.io/realtime/signatures"
)

func main() {
	es := realtime.EventServer{}
	es.SetSignatures(signatures.Signatures)

	// Development switch, processes a onemon.pb file.
	if len(os.Args) == 2 {
		es.OnemonReaderPath(os.Args[1])
		return
	}

	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<addr> <cwd>")
	}

	es.Connect(os.Args[1])
	es.SetCwd(os.Args[2])
	es.Subscribe("massurltask")

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
