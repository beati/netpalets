package main

import (
	//"fmt"
	"bytes"
	"encoding/binary"
	"github.com/beati/netpalets/gamestate"
	"github.com/beati/netpalets/rendering"
	"github.com/beati/netpalets/rtgp"
	"github.com/beati/netpalets/sdl"
	"log"
	"runtime"
	"time"
)

func main() {
	runtime.LockOSThread()
	//runtime.GOMAXPROCS(4)
	var err error

	err = sdl.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer sdl.Quit()

	gameState := gamestate.NewGameState()

	rendering.InitRendering()
	defer rendering.CloseRendering()

	//sdl.ShowCursor(false)

	msgTypes := make([]rtgp.MsgType, 2)
	msgTypes[0] = rtgp.MsgType{128, false}
	msgTypes[1] = rtgp.MsgType{8, true}
	c, err := rtgp.NewConn(":3001", msgTypes, 100)
	if err != nil {
		log.Fatal(err)
	}
	err = c.SetRemoteAddrAndSessionID("195.154.73.145:3000", 0)
	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()

	gc := make(chan []byte)
	go func() {
		for {
			_, g := c.RecvMsg()
			gc <- g
		}
	}()

	g := <-gc
	for sdl.Running {
		select {
		case g = <-gc:
		default:
		}
		rendering.RenderFromNet(g)

		sdl.HandleEvents()
		if sdl.Mouse.Down {
			gameState.Launch(0, sdl.Mouse.X, sdl.Mouse.Y)
			var in bytes.Buffer
			x := int32(sdl.Mouse.X)
			y := int32(sdl.Mouse.Y)
			binary.Write(&in, binary.LittleEndian, x)
			binary.Write(&in, binary.LittleEndian, y)
			c.SendReliableMsg(1, in.Bytes(), false)
		}

		dt := time.Since(t)
		t = time.Now()
		gameState.Step(dt)
	}
}
