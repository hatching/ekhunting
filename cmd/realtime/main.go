// Copyright (C) 2018-2019 Hatching B.V.
// All rights reserved.

package main

import (
	"log"
	"os"
	"time"

	"hatching.io/ekhunting/realtime"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<addr> <cwd>")
	}

	addr := os.Args[1]
	cwd := os.Args[2]

	es := realtime.EventServer{}
	es.SetCwd(cwd)
	es.Connect(addr)
	es.Subscribe("massurltask", "tosti")

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
