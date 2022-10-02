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
	//"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ColdWaterIndex  = 5
	HotWaterIndex   = 6
	HeaterLoopIndex = 7
)

type Logic struct {
	arduino     *ArduinoIoBoard
	tempSensors *TempSensors

	controlBitMask int

	domesticHeatUntil   time.Time
	downstairsHeatUntil time.Time
	upstairsHeatUntil   time.Time

	wholeHouseFan      OnOffThing
	heaterLoopPump     OnOffThing
	recircDHPump       OnOffThing
	downstairsHeatPump OnOffThing
	upstairsHeatPump   OnOffThing

	last *ArduinoBoardStatus

	relayMutex      sync.Mutex
	downstairsMutex sync.Mutex
	wg              sync.WaitGroup
	done            chan bool

	downstairsTemp float64

	// Metrics
	coldWaterCounter    prometheus.Counter
	hotWaterCounter     prometheus.Counter
	heaterLoopCounter   prometheus.Counter
	changeCounter       prometheus.Counter
	downstairsTempGauge prometheus.Gauge
}

func NewLogic(arduino *ArduinoIoBoard, ts *TempSensors) *Logic {
	l := &Logic{
		arduino:     arduino,
		tempSensors: ts,
		coldWaterCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: "physical",
			Name:      "cold_water_usage",
			Help:      "cold water usage counter in gallons",
		}),
		hotWaterCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: "physical",
			Name:      "hot_water_usage",
			Help:      "hot water usage counter in gallons",
		}),
		heaterLoopCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: "physical",
			Name:      "heater_loop_flow",
			Help:      "heater loop flow counter in gallons",
		}),
		changeCounter: promauto.NewCounter(prometheus.CounterOpts{
			Subsystem: "physical",
			Name:      "update_count",
			Help:      "the count of the updates",
		}),
		downstairsTempGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Subsystem: "physical",
			Name:      "downstairs_target_temp",
			Help:      "the target temperature for downstairs (F)",
		}),
	}

	l.wholeHouseFan = NewOnOffThing(OnOffThingOpts{
		Namespace: "heaticus_maximus",
		Name:      "whole_house_fan",
		Gpio: func(on bool) {
			l.control(32, on)
		},
	})

	l.heaterLoopPump = NewOnOffThing(OnOffThingOpts{
		Namespace: "heaticus_maximus",
		Name:      "heater_loop_pump",
		Gpio: func(on bool) {
			l.control(1, on)
		},
	})

	l.recircDHPump = NewOnOffThing(OnOffThingOpts{
		Namespace:      "heaticus_maximus",
		Name:           "recirculating_domestic_hot_pump",
		BlackoutPeriod: time.Minute * 0,
		Gpio: func(on bool) {
			l.control(2, on)
		},
	})

	l.downstairsHeatPump = NewOnOffThing(OnOffThingOpts{
		Namespace: "heaticus_maximus",
		Name:      "downstairs_heat_pump",
		Gpio: func(on bool) {
			l.control(8, on)
		},
	})

	l.upstairsHeatPump = NewOnOffThing(OnOffThingOpts{
		Namespace: "heaticus_maximus",
		Name:      "upstairs_heat_pump",
		Gpio: func(on bool) {
			l.control(4, on)
		},
	})

	if nil != ts {
		l.wg.Add(1)
		go l.downstairsThermostat()
	}

	return l
}

func (l *Logic) Start() (err error) {
	l.arduino.Update = l.Update

	err = l.arduino.Open()
	if nil != err {
		//return err
	}

	return nil
}

func (l *Logic) Stop() {
	l.wholeHouseFan.Shutdown()
	l.heaterLoopPump.Shutdown()
	l.recircDHPump.Shutdown()
	l.downstairsHeatPump.Shutdown()
	l.upstairsHeatPump.Shutdown()
	l.done <- true
	l.wg.Wait()
}

func (l *Logic) Preheat() {
	l.heaterLoopPump.NeededUntil("domestic", time.Now().Add(time.Minute*3))
	l.recircDHPump.OnUntil(time.Now().Add(time.Minute * 3))
}

func (l *Logic) Fan(until time.Time) {
	l.wholeHouseFan.OnUntil(until)
}

func (l *Logic) HeatUpstairs(until time.Time) {
	l.heaterLoopPump.NeededUntil("upstairs", until)
	l.upstairsHeatPump.OnUntil(until)
}

func (l *Logic) HeatDownstairs(until time.Time) {
	l.heaterLoopPump.NeededUntil("downstairs", until)
	l.downstairsHeatPump.OnUntil(until)
}

func (l *Logic) SetDownstairsTarget(goal float64) {
	l.downstairsMutex.Lock()
	l.downstairsTemp = goal
	l.downstairsMutex.Unlock()

	l.downstairsTempGauge.Set(goal)
}

func (l *Logic) Update(s *ArduinoBoardStatus) {
	//fmt.Printf("Update!\n")
	l.changeCounter.Inc()

	if nil == l.last {
		l.last = s
	}

	if s.Inputs[ColdWaterIndex] != l.last.Inputs[ColdWaterIndex] {
		/* Cold water has increased 0.1G */
		//fmt.Printf("Cold++\n")
		l.coldWaterCounter.Add(0.1)
	}
	if s.Inputs[HotWaterIndex] != l.last.Inputs[HotWaterIndex] {
		/* Hot water has increased 0.1G */
		//fmt.Printf("Hot++\n")
		l.hotWaterCounter.Add(0.1)

		/* Make hot water because we know we need it. */
		l.heaterLoopPump.NeededUntil("domestic", time.Now().Add(time.Second*30))
		//l.recircDHPump.OnUntil(time.Now().Add(time.Second * 30))
	}
	if s.Inputs[HeaterLoopIndex] != l.last.Inputs[HeaterLoopIndex] {
		/* Heater Loop has increased 0.1G */
		//fmt.Printf("Heater++\n")
		l.heaterLoopCounter.Add(0.1)
	}
	l.last = s
}

func (l *Logic) control(bit int, on bool) {
	l.relayMutex.Lock()
	defer l.relayMutex.Unlock()

	if on {
		l.controlBitMask |= bit
	} else {
		l.controlBitMask &^= bit
	}

	l.arduino.SetRelayState(l.controlBitMask)
}

func (l *Logic) downstairsThermostat() {
	defer l.wg.Done()

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-l.done:
			t.Stop()
			return
		case <-t.C:
			present := (*l.tempSensors).Get("downstairs_main")
			l.downstairsMutex.Lock()
			target := l.downstairsTemp
			l.downstairsMutex.Unlock()

			if present < target {
				l.HeatDownstairs(time.Now().Add(time.Minute * 3))
				//l.HeatUpstairs(time.Now().Add(time.Minute * 3))
			}
		}
	}
}
