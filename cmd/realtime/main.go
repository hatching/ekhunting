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
	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<addr> <cwd>")
	}

	es := realtime.New(os.Args[2], signatures.Signatures)

	wait := make(chan int)

	es.Connect(os.Args[1])
	es.Subscribe("massurltask", "dumptls", "longtermtask")

	<-wait
}
