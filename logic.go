// Copyright 2019 Weston Schmidt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	ColdWaterIndex  = "5"
	HotWaterIndex   = "6"
	HeaterLoopIndex = "7"
)

type Logic struct {
	Arduino *ArduinoIoBoard

	wholeHouseFan      bool
	heaterLoopPump     bool
	recircDHPump       bool
	downstairsHeatPump bool
	upstairsHeatPump   bool

	last *ArduinoBoardStatus

	heaterLoopUntil time.Time
	ticker          *time.Ticker
	done            chan bool
	wg              sync.WaitGroup
	mutex           sync.Mutex
}

func (l *Logic) Start() (err error) {
	l.Arduino.Update = l.Update

	l.ticker = time.NewTicker(time.Second)

	err = l.Arduino.Open()
	if nil == err {
		l.wg.Add(1)
		go l.run()
	}

	return err
}

func (l *Logic) Stop() {
	l.done <- true
	l.wg.Wait()
}

func (l *Logic) Update(s *ArduinoBoardStatus) {
	fmt.Printf("Update!\n")
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if nil != l.last {
		last := l.last
		if s.Inputs[ColdWaterIndex].State != last.Inputs[ColdWaterIndex].State {
			/* Cold water has increased 0.1G */
			fmt.Printf("Cold++\n")

			/* Make hot water because we think we'll need it. */
			l.heaterLoopPump = true
			l.heaterLoopUntil = time.Now().Add(time.Second * 30)
		}
		if s.Inputs[HotWaterIndex].State != last.Inputs[HotWaterIndex].State {
			/* Hot water has increased 0.1G */
			fmt.Printf("Hot++\n")

			/* Make hot water because we know we need it. */
			l.heaterLoopPump = true
			l.heaterLoopUntil = time.Now().Add(time.Second * 30)
		}
		if s.Inputs[HeaterLoopIndex].State != last.Inputs[HeaterLoopIndex].State {
			/* Heater Loop has increased 0.1G */
			fmt.Printf("Heater++\n")
		}
		l.pushRelayState()
	}
	l.last = s
}

func (l *Logic) pushRelayState() {
	l.Arduino.SetRelayState(l.getRelayState())
}

func (l *Logic) getRelayState() (rv int) {
	if l.heaterLoopPump {
		rv |= 1
	}
	if l.recircDHPump {
		rv |= 2
	}
	if l.upstairsHeatPump {
		rv |= 4
	}
	if l.downstairsHeatPump {
		rv |= 8
	}
	// 16 is unused
	if l.wholeHouseFan {
		rv |= 32
	}

	fmt.Printf("Setting: 0x%02x\n", rv)
	return rv
}

func (l *Logic) run() {
	for {
		select {
		case <-l.done:
			fmt.Printf("Done!\n")
			l.ticker.Stop()
			return
		case <-l.ticker.C:
			l.mutex.Lock()
			if time.Now().After(l.heaterLoopUntil) {
				fmt.Printf("Stop Heater!\n")
				l.heaterLoopPump = false
				l.pushRelayState()
			}
			l.mutex.Unlock()
		}
	}
}
