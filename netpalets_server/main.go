package main

import (
	"bytes"
	//"fmt"
	"github.com/beati/netpalets/gamestate"
	"github.com/beati/netpalets/rtgp"
	"log"
	"time"
)

func play(dataLock chan []byte) {
	g := gamestate.NewGameState()
	g.Launch(0, 1, 1)
	ticker := time.NewTicker(15 * time.Millisecond)
	t := time.Now()
	for {
		select {
		case <-ticker.C:
		}

		dt := time.Since(t)
		t = time.Now()
		g.Step(dt)
		<-dataLock
		var b bytes.Buffer
		g.Serialize(&b)
		dataLock <- b.Bytes()
	}
}

func main() {

	msgTypes := []rtgp.MsgType{rtgp.MsgType{128, false}}
	c1, err := rtgp.NewConn(":3000", msgTypes, 100)
	if err != nil {
		log.Fatal(err)
	}
	err = c1.SetRemoteAddrAndSessionID("127.0.0.1:3001", 1)
	if err != nil {
		log.Fatal(err)
	}
	c2, err := rtgp.NewConn(":3000", msgTypes, 30)
	if err != nil {
		log.Fatal(err)
	}
	err = c2.SetRemoteAddrAndSessionID("127.0.0.1:3002", 2)
	if err != nil {
		log.Fatal(err)
	}

	dataLock := make(chan []byte, 1)
	dataLock <- make([]byte, msgTypes[0].Size)
	go play(dataLock)
	c1.SendPeriodicMsg(0, dataLock)
	c2.SendPeriodicMsg(0, dataLock)

	l := make(chan bool)
	<-l
}
