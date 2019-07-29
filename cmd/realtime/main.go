// Copyright (C) 2019 Hatching B.V.
// All rights reserved.

package main

import (
	"log"
	"os"

	"github.com/hatching/ekhunting/realtime"
	"github.com/hatching/ekhunting/realtime/signatures"
)

func main() {
	es := realtime.EventServer{}
	es.InitApps()
	es.SetSignatures(signatures.Signatures)

	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<addr> <cwd>")
	}

	wait := make(chan int)

	es.Connect(os.Args[1])
	es.SetCwd(os.Args[2])
	es.Subscribe("massurltask", "dumptls", "longtermtask")

	<-wait
}
