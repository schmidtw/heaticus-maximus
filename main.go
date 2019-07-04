package main

import (
	"fmt"
	"time"
)

func main() {

	fmt.Printf("Hi\n")

	names, _ := FindArduinos()
	a := &ArduinoIoBoard{
		Filename: names[0]}
	l := &Logic{Arduino: a}
	l.Start()

	for {
		time.Sleep(time.Second)
	}
}
