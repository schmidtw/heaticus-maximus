// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package gpio

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/tca95xx"
)

func TestNew(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		expectErr   error
	}{
		{
			description: "basic test",
			config: Config{
				InputI2CAddress:   0x20,
				OutputI2CAddress:  0x23,
				InputSamplingRate: 100 * physic.Hertz,
				DebounceTime:      20 * time.Millisecond,
				DebouncedInputs: []DebouncedInput{
					{
						Input:    1,
						Listener: nil,
					},
				},
			},
		}, {
			description: "Pass invalid sampling rate",
			config: Config{
				InputI2CAddress:   0x20,
				OutputI2CAddress:  0x23,
				InputSamplingRate: 100000 * physic.Hertz, // too fast for i2c
				DebounceTime:      20 * time.Millisecond,
				DebouncedInputs: []DebouncedInput{
					{
						Input:    1,
						Listener: nil,
					},
				},
			},
			expectErr: errSampleRateTooFast,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)

			g, err := New(tc.config)

			if tc.expectErr == nil {
				assert.NotNil(g)
				assert.NoError(err)
				return
			}

			assert.ErrorIs(err, tc.expectErr)
			assert.Nil(g)
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		description string
		config      Config
		cancelSet   bool
		openErr     bool
		connectErr  int
		expectErr   error
	}{
		{
			description: "basic test",
			config: Config{
				I2cFile:           "/fake/i2c",
				InputI2CAddress:   0x20,
				OutputI2CAddress:  0x23,
				InputSamplingRate: 1 * physic.Hertz,
				DebounceTime:      20 * time.Millisecond,
				DebouncedInputs: []DebouncedInput{
					{
						Input:    1,
						Listener: nil,
					},
				},
			},
		}, {
			description: "fails trying to open already open connection",
			cancelSet:   true,
			expectErr:   errAlreadyStarted,
		}, {
			description: "open fails",
			openErr:     true,
			expectErr:   errAlreadyStarted,
		}, {
			description: "1st connect fails",
			connectErr:  1,
			expectErr:   errAlreadyStarted,
		}, {
			description: "2nd connect fails",
			connectErr:  2,
			expectErr:   errAlreadyStarted,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			g, err := New(tc.config)
			require.NotNil(g)
			require.NoError(err)

			w := new(mockWrapper)

			if tc.cancelSet {
				_, g.cancel = context.WithCancel(context.Background())
			} else {
				if tc.openErr {
					w.On("Open", mock.Anything).Return(tc.expectErr).Once()
				} else {
					w.On("Open", mock.Anything).Return(nil).Once()
					w.On("Close").Return(nil).Once()
				}

				c0 := new(mockConn)
				c1 := new(mockConn)
				p0 := new(mockPinIO)
				p1 := new(mockPinIO)
				p2 := new(mockPinIO)
				p3 := new(mockPinIO)
				c := []conn.Conn{c0, c1}
				p := [][]tca95xx.Pin{
					{p0, p1},
					{p2, p3},
				}

				switch tc.connectErr {
				case 0:
					w.On("Connect", mock.Anything, mock.Anything).Return(c, p, nil).Once()
					w.On("Close").Return(nil).Once()
					w.On("Connect", mock.Anything, mock.Anything).Return(c, p, nil).Once()
					w.On("Close").Return(nil).Once()
				case 1:
					w.On("Connect", mock.Anything, mock.Anything).Return([]conn.Conn{}, [][]tca95xx.Pin{}, tc.expectErr).Once()
				case 2:
					w.On("Connect", mock.Anything, mock.Anything).Return(c, p, nil).Once()
					w.On("Close").Return(nil).Once()
					w.On("Connect", mock.Anything, mock.Anything).Return([]conn.Conn{}, [][]tca95xx.Pin{}, tc.expectErr).Once()
				}
			}
			g.ioWrapper = w

			ctx := context.Background()

			err = g.Start(ctx)
			if tc.expectErr != nil {
				assert.ErrorIs(err, tc.expectErr)
				return
			}

			assert.NoError(err)

			g.Stop(ctx)
		})
	}
}
