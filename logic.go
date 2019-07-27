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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ColdWaterIndex  = "5"
	HotWaterIndex   = "6"
	HeaterLoopIndex = "7"
)

type Logic struct {
	arduino *ArduinoIoBoard

	wholeHouseFan      bool
	heaterLoopPump     bool
	recircDHPump       bool
	downstairsHeatPump bool
	upstairsHeatPump   bool

	last *ArduinoBoardStatus

	heaterLoopUntil time.Time
	recircUntil     time.Time
	recircBlock     time.Time
	fanUntil        time.Time
	ticker          *time.Ticker
	done            chan bool
	wg              sync.WaitGroup
	mutex           sync.Mutex

	// Metrics
	whFanGauge  prometheus.Gauge
	whFanOnTime prometheus.Counter
}

func NewLogic(arduino *ArduinoIoBoard) *Logic {
	l := &Logic{
		arduino: arduino,
		done:    make(chan bool),
		whFanGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Subsystem: "physical",
			Name:      "wholehouse_fan_state",
			Help:      "Wholehouse fan state (off=0, on=1) at the moment.",
		}),
		whFanOnTime: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: "physical",
			Name:      "wholehouse_fan_on_time",
			Help:      "Wholehouse fan total on time in seconds.",
		}),
	}

	return l
}

func (l *Logic) Start() (err error) {
	l.arduino.Update = l.Update

	l.ticker = time.NewTicker(time.Second)

	err = l.arduino.Open()
	if nil != err {
		//return err
	}

	l.wg.Add(1)
	go l.run()

	return nil
}

func (l *Logic) Stop() {
	l.done <- true
	l.wg.Wait()
}

func (l *Logic) Preheat() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.preheat(true)
	l.pushRelayState()
}

func (l *Logic) preheat(force bool) {
	if force || (false == l.heaterLoopPump && time.Now().After(l.recircBlock)) {
		l.recircUntil = time.Now().Add(time.Second * 30)
		l.recircBlock = time.Now().Add(time.Minute * 5)
		l.recircDHPump = true
	}
}

func (l *Logic) Fan(until time.Time) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.whFanGauge.Set(1.0)
	l.wholeHouseFan = true
	l.fanUntil = until
	l.pushRelayState()
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
		}
		if s.Inputs[HotWaterIndex].State != last.Inputs[HotWaterIndex].State {
			/* Hot water has increased 0.1G */
			fmt.Printf("Hot++\n")

			/* Make hot water because we know we need it. */
			l.heaterLoopPump = true
			l.heaterLoopUntil = time.Now().Add(time.Second * 30)
			l.preheat(false)
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
	l.arduino.SetRelayState(l.getRelayState())
	fmt.Printf("Done pushing\n")
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
	defer l.wg.Done()
	for {
		select {
		case <-l.done:
			fmt.Printf("Done!\n")
			l.ticker.Stop()
			return
		case <-l.ticker.C:
			l.mutex.Lock()
			now := time.Now()
			push := false

			if l.wholeHouseFan {
				l.whFanOnTime.Inc()
				if now.After(l.fanUntil) {
					l.whFanGauge.Set(0.0)
					l.wholeHouseFan = false
					push = true
				}
			}

			if l.recircDHPump && now.After(l.recircUntil) {
				l.recircDHPump = false
				push = true
			}

			if l.heaterLoopPump && now.After(l.heaterLoopUntil) {
				l.heaterLoopPump = false
				push = true
			}

			if push {
				l.pushRelayState()
			}

			l.mutex.Unlock()
		}
	}
}
