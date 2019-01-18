// Copyright (C) 2019 Hatching B.V.
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

	es := realtime.EventServer{}
	es.Connect(os.Args[1])
	es.SetCwd(os.Args[2])
	es.Subscribe("massurltask")

	for {
		time.Sleep(100 * time.Millisecond)
	}
}
