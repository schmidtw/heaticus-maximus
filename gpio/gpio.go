// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package gpio

import (
	"context"
	"errors"
	"sync"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/tca95xx"
)

var (
	errSampleRateTooFast = errors.New("sample rate too fast")
	errAlreadyStarted    = errors.New("already started")
)

var (
	// This is for SequentMicrosystems.com 8-Relays v5.0 board.
	relayToBitMap = map[int]int{
		1: 0,
		8: 1,
		2: 2,
		7: 3,
		4: 4,
		5: 5,
		3: 6,
		6: 7,
	}

	// This is for SequentMicrosystems.com 16 Opto-Isolated Inputs Hat v1.0 board.
	inputToBitMap = map[int]inputPinPortMap{
		16: {port: 0, bit: 0},
		15: {port: 0, bit: 1},
		14: {port: 0, bit: 2},
		13: {port: 0, bit: 3},
		12: {port: 0, bit: 4},
		11: {port: 0, bit: 5},
		10: {port: 0, bit: 6},
		9:  {port: 0, bit: 7},
		8:  {port: 1, bit: 0},
		7:  {port: 1, bit: 1},
		6:  {port: 1, bit: 2},
		5:  {port: 1, bit: 3},
		4:  {port: 1, bit: 4},
		3:  {port: 1, bit: 5},
		2:  {port: 1, bit: 6},
		1:  {port: 1, bit: 7},
	}
)

type inputPinPortMap struct {
	port int
	bit  int
}

type DebouncedState struct {
	Input int
	State int // 0 or 1
}

type DebouncedInput struct {
	Input    int                   // Input number from the board
	Listener chan<- DebouncedState // The channel to notifiy on state change
}

type Config struct {
	I2cFile string

	InputI2CAddress   int
	OutputI2CAddress  int
	InputSamplingRate physic.Frequency
	DebounceTime      time.Duration
	DebouncedInputs   []DebouncedInput
}

type Gpio struct {
	m         sync.Mutex
	config    Config
	cancel    context.CancelFunc
	readPorts []bool

	ioWrapper gpioWrapper
	in        []conn.Conn
	out       [][]tca95xx.Pin
	wg        sync.WaitGroup
}

type gpioWrapper interface {
	Open(string) error
	Close() error
	Connect(tca95xx.Variant, int) ([]conn.Conn, [][]tca95xx.Pin, error)
}

func (c Config) toPorts() []bool {
	out := make([]bool, 2)

	for _, input := range c.DebouncedInputs {
		port := inputToBitMap[input.Input]
		out[port.port] = true
	}

	return out
}

func New(c Config) (*Gpio, error) {
	if c.InputSamplingRate > physic.Hertz*10000 {
		return nil, errSampleRateTooFast
	}

	return &Gpio{
		config:    c,
		readPorts: c.toPorts(),
		ioWrapper: &hwWrapper{},
	}, nil
}

func (g *Gpio) Start(ctx context.Context) (err error) {
	g.m.Lock()
	defer g.m.Unlock()

	if g.cancel != nil {
		return errAlreadyStarted
	}

	if err := g.ioWrapper.Open(g.config.I2cFile); err != nil {
		return err
	}

	g.in, _, err = g.ioWrapper.Connect(tca95xx.TCA9535, g.config.InputI2CAddress)
	if err != nil {
		_ = g.ioWrapper.Close()
		return err
	}

	_, g.out, err = g.ioWrapper.Connect(tca95xx.TCA9534, g.config.OutputI2CAddress)
	if err != nil {
		_ = g.ioWrapper.Close()
		return err
	}

	ctx, g.cancel = context.WithCancel(ctx)
	g.wg.Add(1)
	go g.loop(ctx)

	return nil
}

func (g *Gpio) Stop(ctx context.Context) {
	g.m.Lock()
	defer g.m.Unlock()

	if g.cancel != nil {
		g.cancel()
		g.wg.Wait()
		g.cancel = nil
	}

	_ = g.ioWrapper.Close()
}

func (g *Gpio) loop(ctx context.Context) {
	sampleTicker := time.NewTicker(g.config.InputSamplingRate.Period())

	ignoreUntil := make(map[int]time.Time, len(g.config.DebouncedInputs))
	current := make(map[int]int, len(g.config.DebouncedInputs))

	for {
		select {
		case <-sampleTicker.C:
			g.readAndDebounce(ignoreUntil, current)
		case <-ctx.Done():
			g.wg.Done()
			return
		}
	}
}

func (g *Gpio) ReadInputs() (out map[int]int, err error) {
	g.m.Lock()
	defer g.m.Unlock()

	rx := make([]byte, 2)

	for i, p := range g.readPorts {
		if p {
			r := make([]byte, 1)
			if err := g.in[i].Tx(nil, r); err != nil {
				return nil, err
			}
			rx[i] = r[0]
		}
	}

	for _, di := range g.config.DebouncedInputs {
		bitMap := inputToBitMap[di.Input]

		out[di.Input] = int(rx[bitMap.port] & (1 << bitMap.bit))
	}

	return out, nil
}

func (g *Gpio) readAndDebounce(ignoreUntil map[int]time.Time, vals map[int]int) {
	now := time.Now()
	newVals, err := g.ReadInputs()
	if err != nil {
		return
	}

	for k, v := range newVals {
		if vals[k] != v {
			if ignoreUntil[k].Before(now) {
				vals[k] = v
				ignoreUntil[k] = now.Add(g.config.DebounceTime)

				g.config.DebouncedInputs[k].Listener <- DebouncedState{
					Input: k,
					State: v,
				}
			}
		}
	}
}
