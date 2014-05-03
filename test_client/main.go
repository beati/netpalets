package main

import (
	"bytes"
	"encoding/binary"
	//"fmt"
	"github.com/beati/netpalets/sdl"
	"github.com/beati/netpalets/rtgp"
	"log"
	"runtime"
)

func shiftPos(x float64) int {
	return int(x+0.5) - 25
}

type vector struct {
	X float64
	Y float64
}

type pos struct {
	X uint32
	Y uint32
}

func main() {
	runtime.LockOSThread()

	var err error
	err = sdl.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Test client", 640, 480)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyWindow(window)

	renderer, err := sdl.CreateRenderer(window, -1)
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyRenderer(renderer)

	paletGfx, err := sdl.LoadBMP(renderer, "palet.bmp")
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.DestroyTexture(paletGfx)

	PayloadTypes := []rtgp.PayloadType{{20, true}, {8, false}}
	conn, err := rtgp.NewConn(":3002", PayloadTypes)
	conn.SetTickRate(100)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetRemoteAddrAndSessionId("127.0.0.1:3001", 1)

	palet := vector{320, 240}

	time := uint32(0)

	for sdl.Running {
		sdl.HandleEvents()

		if sdl.Mouse_state.Down {
			var b bytes.Buffer
			x := uint32(sdl.Mouse_state.X)
			y := uint32(sdl.Mouse_state.Y)
			pos := pos{x, y}
			binary.Write(&b, binary.LittleEndian, pos)
			conn.Send(rtgp.Payload{1, b.Bytes()})
		}

		recved := conn.Recv()
		for _, p := range recved {
			if p.Type == 0 {
				var t uint32
				data := bytes.NewReader(p.Data)
				binary.Read(data, binary.LittleEndian, &t)
				if t >= time {
					time = t
					binary.Read(data, binary.LittleEndian,
						&palet)
				}
			}
		}

		err = sdl.RenderClear(renderer)
		if err != nil {
			log.Fatal(err)
		}
		err = sdl.RenderCopy(renderer, paletGfx, shiftPos(palet.X),
			shiftPos(palet.Y), 50, 50)
		if err != nil {
			log.Fatal(err)
		}
		sdl.RenderPresent(renderer)
		//select {}
	}
}
