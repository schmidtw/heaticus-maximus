// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package watermeter

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/schmidtw/heaticus-maximus/units"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

// Config provides the watermeter configuration options.
type Config struct {
	VolumePerPulse units.Volume `yaml:"volume_per_pulse"`
	StartingVolume units.Volume `yaml:"staring_volume"`
	MaxEventCount  int          `yaml:"max_event_count"`
}

type Option interface {
	apply(w *Watermeter)
}

type Watermeter struct {
	name          string
	mutex         sync.Mutex
	clock         clock.Clock
	total         units.Volume
	pulse         units.Volume
	maxEventCount int
	events        list.List
}

// New makes a new watermeter.
func New(name string, cfg Config, opts ...Option) (*Watermeter, error) {
	if cfg.MaxEventCount < 1 {
		cfg.MaxEventCount = 100
	}
	if cfg.VolumePerPulse <= 0.0 {
		return nil, ErrInvalidParameter
	}
	if cfg.StartingVolume <= 0.0 {
		cfg.StartingVolume = 0.0
	}

	w := Watermeter{
		name:          name,
		clock:         clock.New(),
		total:         cfg.StartingVolume,
		pulse:         cfg.VolumePerPulse,
		maxEventCount: cfg.MaxEventCount,
	}

	w.events.Init()

	for _, opt := range opts {
		opt.apply(&w)
	}

	return &w, nil
}

func (w *Watermeter) Pulse() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.total += w.pulse

	now := w.clock.Now()
	w.events.PushFront(now)

	for w.events.Len() > w.maxEventCount {
		w.events.Remove(w.events.Back())
	}
}

func (w Watermeter) Flow(over time.Duration) units.VolumeFlowRate {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	until := w.clock.Now().Add(-1 * over)

	var pulses int
	for event := w.events.Front(); event.Next() != nil; event = event.Next() {
		t := event.Value.(time.Time)
		if until.Before(t) {
			pulses++
		}
	}
	rate := (float64(pulses) * float64(w.pulse)) / over.Minutes()

	return units.VolumeFlowRate(rate)
}

func (w Watermeter) Total() units.Volume {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.total
}

func (w Watermeter) String() string {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.name + ":" + w.total.String()
}

// UseClock provides a way to set the clock used.  This is used for testing.
func UseClock(c clock.Clock) Option {
	return &clockOption{clk: c}
}

type clockOption struct {
	clk clock.Clock
}

func (c clockOption) apply(w *Watermeter) {
	w.clock = c.clk
}
