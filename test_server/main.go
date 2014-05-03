package main

import (
	"bytes"
	"encoding/binary"
	//"fmt"
	"github.com/beati/netpalets/rtgp"
	"log"
)

type vector struct {
	X float64
	Y float64
}

type pos struct {
	X uint32
	Y uint32
}

func main() {
	PayloadTypes := []rtgp.PayloadType{{20, true}, {8, false}}
	rtgp.RegisterPayloadTypes(PayloadTypes)
	rtgp.SetTickRate(100)
	conn, err := rtgp.NewConn(":3001")
	if err != nil {
		log.Fatal(err)
	}
	conn.SetRemoteAddrAndSessionId("127.0.0.1:3002", 2)

	time := uint32(0)
	palet := vector{320, 240}

	var b bytes.Buffer
	binary.Write(&b, binary.LittleEndian, time)
	binary.Write(&b, binary.LittleEndian, palet)
	dataLock := make(chan rtgp.Payload, 1)
	dataLock <- rtgp.Payload{0, b.Bytes()}
	conn.SendPeriodically(dataLock)

	for {
		recved := conn.Recv()
		for _, p := range recved {
			if p.Type == 1 {
				var pos pos
				data := bytes.NewReader(p.Data)
				binary.Read(data, binary.LittleEndian, &pos)
				palet.X = float64(pos.X)
				palet.Y = float64(pos.Y)
			}
		}

		var b bytes.Buffer
		binary.Write(&b, binary.LittleEndian, time)
		binary.Write(&b, binary.LittleEndian, palet)
		<-dataLock
		dataLock <- rtgp.Payload{0, b.Bytes()}

		time++
	}
}
