package fatal

import (
	"log"
	"runtime/debug"
)

func Check(err error) {
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
}
