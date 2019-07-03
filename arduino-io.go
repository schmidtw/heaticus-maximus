package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/schmidtw/go-serial"
)

func FindArduino() (string, error) {
	list, err := serial.FindSerialPorts()
	if nil == err {
		for _, v := range list {
			if rv, _ := regexp.MatchString(".*Arduino.*", v); true == rv {
				fmt.Printf("Found\n")
				return v, err
			}
		}
	}
	return "", err
}

type ArduinoIoBoard struct {
	Name         string
	SerialNumber string
	serial       *serial.Serial
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
		return fmt.Errorf("Arduino '%s' already open.", a.Name)
	}

	a.serial = &serial.Serial{
		Name:      a.Name,
		Baud:      115200,
		Config:    "8N1",
		Canonical: false,
		Vmin:      200,
		Vtime:     1,
	}

	if err = a.serial.Open(); nil == err {
		a.read()
		/*
			s, _ := a.read()
			d := json.NewDecoder(strings.NewReader(s))
			var status ArduinoBoardStatus
			d.Decode(&status)

			fmt.Printf("got: '%s'\n\n%#v\n", s, status)
		*/
	} else {
		fmt.Printf("err %#v\n", err)
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

/*
	if _, err := os.Stat("/dev/serial"); nil == err {
		if _, err := os.Stat("/dev/serial/by-id"); nil == err {
		}
	}
*/

/*
	s := &serial.Serial{Name: "/dev/ttyACM0"}
	s.Open()
	s.SetBaud(119200, "8N1")
	b := make([]byte, 1000)
	n, _ := s.Read(b)
	fmt.Printf("%d - %s\n", n, string(b[:n]))
	for {
		n, _ = s.Read(b)
		fmt.Printf("%d - %s\n", n, string(b[:n]))
	}
	s.Close()
			fmt.Printf("%d - %s\n", n, string(b[:n]))

		ports, err := serial.GetPortsList()
		if err != nil {
			fmt.Println(err)
		}
		if len(ports) == 0 {
			fmt.Println("No serial ports found!")
		}
		for _, port := range ports {
			fmt.Printf("Found port: %v\n", port)
		}

		mode := &serial.Mode{
			BaudRate: 115200,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}
		port, err := serial.Open("/dev/ttyACM0", mode)
		if err != nil {
			fmt.Println(err)
		}

		err = port.SetMode(mode)
		if err != nil {
			fmt.Println(err)
		}

		buff := make([]byte, 100)
		for {
			n, err := port.Read(buff)
			if err != nil {
				fmt.Println(err)
				break
			}
			if n == 0 {
				fmt.Println("\nEOF")
				break
			}
			fmt.Printf("%v", string(buff[:n]))
		}
*/
