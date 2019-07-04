package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/schmidtw/go-serial"
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
	quit   chan int
	wg     sync.WaitGroup
}

type ArduinoBoardInputStatus struct {
	PulseCount    int
	State         int
	LastPulseTime float64
}

type ArduinoBoardStatus struct {
	SerialNumber    string
	FirmwareVersion string
	UpTime          int
	RelayState      int
	Inputs          map[string]ArduinoBoardInputStatus `json:-`
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
		Vmin:      200,
		Vtime:     1,
	}

	err = a.serial.Open()

	if nil == err {
		a.wg.Add(1)
		go a.run()
	}
	/*
			a.read()

				s, _ := a.read()
				d := json.NewDecoder(strings.NewReader(s))
				var status ArduinoBoardStatus
				d.Decode(&status)

				fmt.Printf("got: '%s'\n\n%#v\n", s, status)
		} else {
			fmt.Printf("err %#v\n", err)
		}
	*/

	return err
}

func (a *ArduinoIoBoard) Close() {
	if nil != a.serial {
		a.quit <- 0
		a.wg.Wait()
	}
}

func (a *ArduinoIoBoard) run() {

	for {
		s, err := a.read()
		if nil != err {
			var status ArduinoBoardStatus
			d := json.NewDecoder(strings.NewReader(s))
			if nil == d.Decode(&status) {
				if nil != a.Update {
					go a.Update(&status)
				}
			}
		}

		select {
		case <-a.quit:
			break
		}
	}

	a.serial.Close()
	a.wg.Done()
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

func (a *ArduinoIoBoard) Help() (rv string, err error) {
	if nil != a.serial {
		b := []byte{'s', ' ', '0', '\n', '\n'}
		a.serial.Write(b)
		rv, err = a.read()
		if nil == err {
			fmt.Printf("%s\n", rv)
		}
	}

	for {
		rv, err = a.read()
		if nil == err {
			//fmt.Printf("%s\n", rv)
			d := json.NewDecoder(strings.NewReader(rv))
			var status ArduinoBoardStatus
			if nil == d.Decode(&status) {
				fmt.Printf("%d - Out: %d - 5: %d:%d - 6: %d:%d - 7: %d:%d\n", status.UpTime, status.RelayState,
					status.Inputs["5"].State, status.Inputs["5"].PulseCount,
					status.Inputs["6"].State, status.Inputs["6"].PulseCount,
					status.Inputs["7"].State, status.Inputs["7"].PulseCount)
				//fmt.Printf("got: %#v\n", status)
			}
		}
	}

	return rv, err
}

func (a *ArduinoIoBoard) read() (rv string, err error) {
	if nil != a.serial {
		b := make([]byte, 1)
		n := 1
		err = nil
		for 0 < n && nil == err && '\n' != b[0] {
			n, err = a.serial.Read(b)
			rv += string(b[:n])
		}

		if nil != err {
			rv = ""
		}
	}

	return rv, err
}
