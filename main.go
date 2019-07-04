package main

import (
	"fmt"
)

func main() {

	fmt.Printf("Hi\n")

	names, _ := FindArduinos()
	a := &ArduinoIoBoard{Filename: names[0]}
	a.Open()
	a.Help()
}
