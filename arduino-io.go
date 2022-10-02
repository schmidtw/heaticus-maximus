package main

import (
	"fmt"
	"regexp"
	"sync"

	serial "github.com/schmidtw/go232"
)

// FindArduinos looks at the list of serial ports present and returns the list
// of all of them that are Arduinos.
func FindArduinos() (list []string, err error) {
	all, err := serial.FindSerialPorts()
	if nil == err {
		for _, v := range all {
			if rv, _ := regexp.MatchString(".*Arduino.*", v); true == rv {
				list = append(list, v)
			}
		}
	}
	return list, err
}

//
type ArduinoIoBoard struct {
	Filename string
	Update   func(*ArduinoBoardStatus)

	serial *serial.Serial
	done   chan bool
	wg     sync.WaitGroup
}

type ArduinoBoardInputStatus struct {
	State int
}

type ArduinoBoardStatus struct {
	SerialNumber string
	RelayState   int
	Inputs       map[int]int
}

func (a *ArduinoIoBoard) Open() (err error) {
	if nil != a.serial {
		return fmt.Errorf("Arduino '%s' already open.", a.Filename)
	}

	a.serial = &serial.Serial{
		Name:      a.Filename,
		Baud:      115200,
		Config:    "8N1",
		Canonical: false,
		Vmin:      1,
		Vtime:     1,
	}

	err = a.serial.Open()

	if nil == err {
		a.wg.Add(1)
		go a.run()
	}

	return err
}

func (a *ArduinoIoBoard) Close() {
	if nil != a.serial {
		a.done <- true
		a.wg.Wait()
	}
}

func (a *ArduinoIoBoard) run() {

	defer a.serial.Close()
	defer a.wg.Done()

	for {
		s, err := a.read()
		//fmt.Printf("got: '%s'\n", s )
		if nil == err {
			//fmt.Printf("no error: '%s'\n", s )
			var status ArduinoBoardStatus
			var input int
			status.Inputs = make(map[int]int, 8)
			n, err := fmt.Sscanf(s, "%02X|%02X|%02X", &status.SerialNumber, &input, &status.RelayState)
			//fmt.Printf("n = %d, err = %v\n", n, err )
			if nil == err && 3 == n {
				for i := 0; i < 8; i++ {
					status.Inputs[i] = int(1 & (uint(input) >> uint(i)))
					//fmt.Printf( "Inputs[%d] = %d\n", i, status.Inputs[i])
				}
				//fmt.Printf( "Decoding is ok\n")
				if nil != a.Update {
					//fmt.Printf( "Calling Update\n")
					go a.Update(&status)
				}
			}
		}

		select {
		case <-a.done:
			return
		default:
		}
	}
}

func (a *ArduinoIoBoard) SetRelayState(state int) (err error) {
	if nil == a.serial {
		return fmt.Errorf("Arduino '%s' not open.", a.Filename)
	}

	s := fmt.Sprintf("s %d\n", state)
	b := []byte(s)
	for left := len(b); 0 < left; {
		tmp, err := a.serial.Write(b)
		if nil != err {
			return err
		}
		b = b[tmp:]
		left = len(b)
	}
	return err
}

func (a *ArduinoIoBoard) read() (rv string, err error) {
	if nil != a.serial {
		b := make([]byte, 1)
		i := 0
		for i < 9 {
			n, err := a.serial.Read(b)

			if nil != err {
				return "", err
			}

			// The format is always: '00|00|00\n' so if we see a \n before
			// the end, then we're out of sync.  Restart the search.
			var newline byte = '\n'
			if i < 8 && newline == b[0] {
				i = 0
				rv = ""
				continue
			}

			if newline != b[0] {
				rv += string(b[:n])
			}
			i++
		}
	}

	return rv, err
}
