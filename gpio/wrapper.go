// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package gpio

import (
	"fmt"
	"sync"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/tca95xx"
)

type hwWrapper struct {
	m     sync.Mutex
	bus   i2c.BusCloser
	conns []*tca95xx.Dev
}

func (h *hwWrapper) Open(file string) (err error) {
	h.m.Lock()
	defer h.m.Unlock()

	if h.bus != nil {
		return errAlreadyStarted
	}

	h.bus, err = i2creg.Open(file)
	return err
}

func (h *hwWrapper) Close() (err error) {
	h.m.Lock()
	defer h.m.Unlock()

	for i := range h.conns {
		e := h.conns[i].Close()
		if e != nil && err == nil {
			err = e
		}
	}

	e := h.bus.Close()
	if e != nil && err == nil {
		err = e
	}

	return err
}

func (h *hwWrapper) Connect(v tca95xx.Variant, addr int) ([]conn.Conn, [][]tca95xx.Pin, error) {
	h.m.Lock()
	defer h.m.Unlock()

	if h.bus == nil {
		return nil, nil, fmt.Errorf("invalid state")
	}

	dev, err := tca95xx.New(h.bus, v, uint16(addr))
	if err != nil {
		return nil, nil, err
	}

	return dev.Conns, dev.Pins, nil
}
