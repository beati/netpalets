package main

import (
	"bytes"
	//"fmt"
	"encoding/binary"
	"github.com/beati/netpalets/gamestate"
	"github.com/beati/netpalets/rtgp"
	"log"
	"time"
)

type input struct {
	X int32
	Y int32
}

func play(dataLock chan []byte, i1 chan input, i2 chan input) {
	g := gamestate.NewGameState()
	ticker := time.NewTicker(15 * time.Millisecond)
	t := time.Now()
	for {
		select {
		case <-ticker.C:
		case input1 := <-i1:
			g.Launch(0, int(input1.X), int(input1.Y))
		case input2 := <-i2:
			g.Launch(7, int(input2.X), int(input2.Y))
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

func recvInputs(c *rtgp.Conn, i chan input) {
	for {
		_, in := c.RecvMsg()
		r := bytes.NewReader(in)
		var input input
		binary.Read(r, binary.LittleEndian, &input)
		i <- input
	}
}

func main() {
	msgTypes := make([]rtgp.MsgType, 2)
	msgTypes[0] = rtgp.MsgType{128, false}
	msgTypes[1] = rtgp.MsgType{8, true}
	c1, err := rtgp.NewConn(":3000", msgTypes, 100)
	if err != nil {
		log.Fatal(err)
	}
	err = c1.SetRemoteAddrAndSessionID("85.171.104.166:3001", 1)
	if err != nil {
		log.Fatal(err)
	}
	i1 := make(chan input)
	go recvInputs(c1, i1)
	c2, err := rtgp.NewConn(":3000", msgTypes, 100)
	if err != nil {
		log.Fatal(err)
	}
	err = c2.SetRemoteAddrAndSessionID("85.171.104.166:3002", 2)
	if err != nil {
		log.Fatal(err)
	}
	i2 := make(chan input)
	go recvInputs(c2, i2)

	dataLock := make(chan []byte, 1)
	dataLock <- make([]byte, msgTypes[0].Size)
	go play(dataLock, i1, i2)
	c1.SendPeriodicMsg(0, dataLock)
	c2.SendPeriodicMsg(0, dataLock)

	l := make(chan bool)
	<-l
}
