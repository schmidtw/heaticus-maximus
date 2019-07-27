package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {

	fmt.Printf("Hi\n")

	names, _ := FindArduinos()
	a := &ArduinoIoBoard{}
	if nil != names {
		a.Filename = names[0]
	}
	l := NewLogic(a)
	l.Start()

	wh := NewWeb(l, nil)
	wh.Start()

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		wh.Stop()
		l.Stop()
		close(idleConnsClosed)
	}()

	<-idleConnsClosed
}
