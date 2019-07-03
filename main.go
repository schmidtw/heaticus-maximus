package main

import (
	"fmt"
)

func main() {

	fmt.Printf("Hi\n")

	name, _ := FindArduino()
	a := &ArduinoIoBoard{Name: name}
	a.Open()
	a.Help()
}
