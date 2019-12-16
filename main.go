package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {

	fmt.Printf("Hi\n")

	tso := TempSensorsOpts{
		Namespace:    "heaticus_maximus",
		Path:         "/dev/ttyUSB0",
		SamplePeriod: time.Second * 2,
		Names: map[string]string{
			"28.84c5c4331401.5c": "downstairs_main",
		},
	}
	ts, _ := NewTempSensors(tso)
	names, _ := FindArduinos()
	a := &ArduinoIoBoard{}
	if nil != names {
		a.Filename = names[0]
	}
	l := NewLogic(a, &ts)
	l.Start()

	wh := NewWeb(l, &ts, nil)
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
