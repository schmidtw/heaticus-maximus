// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package gpio

import (
	"time"

	"github.com/stretchr/testify/mock"
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/tca95xx"
)

type mockWrapper struct {
	mock.Mock
}

func (m *mockWrapper) Open(file string) (err error) {
	a := m.Called(file)
	return a.Error(0)
}

func (m *mockWrapper) Close() (err error) {
	a := m.Called()
	return a.Error(0)
}

func (m *mockWrapper) Connect(v tca95xx.Variant, addr int) ([]conn.Conn, [][]tca95xx.Pin, error) {
	a := m.Called(v, addr)
	return a.Get(0).([]conn.Conn), a.Get(1).([][]tca95xx.Pin), a.Error(2)
}

// Mocking conn.Conn

type mockConn struct {
	mock.Mock
}

func (m *mockConn) String() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockConn) Tx(w, r []byte) error {
	a := m.Called(w, r)
	return a.Error(0)
}

func (m *mockConn) Duplex() conn.Duplex {
	a := m.Called()
	return a.Get(0).(conn.Duplex)
}

// Mocking tca95xx.Pin

type mockPinIO struct {
	mock.Mock
}

func (m *mockPinIO) Name() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockPinIO) String() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockPinIO) Number() int {
	a := m.Called()
	return a.Int(0)
}

func (m *mockPinIO) Function() string {
	a := m.Called()
	return a.String(0)
}

func (m *mockPinIO) In(pull gpio.Pull, edge gpio.Edge) error {
	a := m.Called(pull, edge)
	return a.Error(0)
}

func (m *mockPinIO) Read() gpio.Level {
	a := m.Called()
	return a.Get(0).(gpio.Level)
}

func (m *mockPinIO) WaitForEdge(timeout time.Duration) bool {
	a := m.Called(timeout)
	return a.Bool(0)
}

func (m *mockPinIO) Pull() gpio.Pull {
	a := m.Called()
	return a.Get(0).(gpio.Pull)
}

func (m *mockPinIO) DefaultPull() gpio.Pull {
	a := m.Called()
	return a.Get(0).(gpio.Pull)
}

func (m *mockPinIO) Out(l gpio.Level) error {
	a := m.Called(l)
	return a.Error(0)
}

func (m *mockPinIO) PWM(duty gpio.Duty, f physic.Frequency) error {
	a := m.Called(duty, f)
	return a.Error(0)
}

func (m *mockPinIO) SetPolarityInverted(p bool) error {
	a := m.Called(p)
	return a.Error(0)
}

func (m *mockPinIO) IsPolarityInverted() (bool, error) {
	a := m.Called()
	return a.Bool(0), a.Error(1)
}

func (m *mockPinIO) Halt() error {
	a := m.Called()
	return a.Error(0)
}
