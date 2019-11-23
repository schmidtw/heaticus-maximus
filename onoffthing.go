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
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type OnOffThing interface {
	// Returns the present state of the thing and when it will transition
	// to a different state
	State() (bool, time.Time)

	// Turns the thing on until the specified time.  If the thing is already
	// on, the time it turns off is adjusted
	OnUntil(time.Time)

	// Turns the thing on until the last thing that needs it expires.
	NeededUntil(string, time.Time)

	// Turns the thing off indefinitely.
	Off()

	// Shuts down and turns everything off.
	Shutdown()
}

type OnOffThingOpts struct {
	// The Namespace of the metrics for the thing
	Namespace string

	// The Name of the thing
	Name string

	// How long after the thing was last turned off before it will turn on
	// the next time.  The duration specified in the OnUntil() call is ignored
	// for this calculation.  The default is no blackout period.
	BlackoutPeriod time.Duration

	Gpio func(on bool)
}

type onOffThing struct {
	name           string
	state          bool
	blackoutPeriod time.Duration
	refreshPeriod  time.Duration
	neededUntil    map[string]time.Time
	until          time.Time
	notBefore      time.Time
	done           chan bool
	wg             sync.WaitGroup
	mutex          sync.Mutex
	refreshTicker  *time.Ticker
	changeTicker   *time.Ticker
	gpio           func(on bool)

	// Metrics
	status prometheus.Gauge
	onTime prometheus.Counter
}

func NewOnOffThing(opts OnOffThingOpts) OnOffThing {
	t := &onOffThing{
		name:           opts.Name,
		done:           make(chan bool),
		neededUntil:    make(map[string]time.Time),
		refreshPeriod:  time.Second,
		blackoutPeriod: opts.BlackoutPeriod,
		gpio:           opts.Gpio,
	}

	if nil == t.gpio {
		t.gpio = func(on bool) {}
	}

	// Initialize the tickers but stop them so we don't need
	// to do nil checks everywhere
	t.refreshTicker = time.NewTicker(time.Minute)
	t.refreshTicker.Stop()
	t.changeTicker = time.NewTicker(time.Minute)
	t.changeTicker.Stop()

	t.status = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: "physical",
		Name:      opts.Name + "_state",
		Help:      opts.Name + " state (0 = off, 1 = on) at this moment.",
	})

	t.onTime = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: "physical",
		Name:      opts.Name + "_on_time",
		Help:      opts.Name + " total on time in seconds.",
	})

	t.wg.Add(1)
	go t.run()

	return t
}

func (t *onOffThing) Shutdown() {
	t.done <- true
	t.wg.Wait()
}

func (t *onOffThing) State() (bool, time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if time.Now().After(t.until) {
		return t.state, time.Time{}
	}
	return t.state, t.until
}

func (t *onOffThing) Off() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.until = time.Now().Add(-1 * time.Nanosecond)
	t.stop()
}

func (t *onOffThing) OnUntil(when time.Time) {
	now := time.Now()
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if now.After(t.notBefore) {
		if false == t.state {
			t.gpio(false)
			t.refreshTicker.Stop()
			t.refreshTicker = time.NewTicker(t.refreshPeriod)
		}
		t.gpio(true)
		t.state = true
		t.until = when
		t.changeTicker.Stop()
		t.changeTicker = time.NewTicker(when.Sub(now))
		t.status.Set(1.0)
	}
}

func (t *onOffThing) NeededUntil(name string, when time.Time) {
	t.neededUntil[name] = when

	until := time.Now()
	for _, v := range t.neededUntil {
		if v.After(until) {
			until = v
		}
	}

	t.OnUntil( until )
}

func (t *onOffThing) run() {
	defer t.wg.Done()
	for {
		t.mutex.Lock()
		select {
		case <-t.done:
			t.stop()
			t.mutex.Unlock()
			return
		case <-t.changeTicker.C:
			t.stop()
			t.mutex.Unlock()

		case <-t.refreshTicker.C:
			t.mutex.Unlock()
			t.onTime.Inc()
		default:
			t.mutex.Unlock()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (t *onOffThing) stop() {
	t.gpio(false)
	t.state = false
	t.notBefore = time.Now().Add(t.blackoutPeriod)
	t.changeTicker.Stop()
	t.refreshTicker.Stop()
	t.status.Set(0.0)
}
