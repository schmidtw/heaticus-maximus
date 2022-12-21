// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package watermeter

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/schmidtw/heaticus-maximus/units"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestNew(t *testing.T) {
	tests := []struct {
		description    string
		cfg            Config
		pulses         int
		pulsePeriod    time.Duration
		expectListSize int
		total          string
		str            string
		over           time.Duration
		after          time.Duration
		flow           string
		expectedErr    error
	}{
		{
			description: "basic test",
			cfg: Config{
				VolumePerPulse: must(units.ParseVolume("1L")),
			},
			pulses:         105,
			pulsePeriod:    time.Minute,
			expectListSize: 100,
			total:          "105.000L",
			str:            "name:105.000L",
			over:           5 * time.Minute,
			flow:           "1.000Lpm",
		}, {
			description: "start with a total, also get back a rate of 0",
			cfg: Config{
				VolumePerPulse: must(units.ParseVolume("1L")),
				StartingVolume: must(units.ParseVolume("1L")),
				MaxEventCount:  10,
			},
			pulses:         15,
			pulsePeriod:    time.Minute,
			expectListSize: 10,
			total:          "16.000L",
			str:            "name:16.000L",
			over:           5 * time.Second,
			after:          5 * time.Minute,
			flow:           "0.000Lpm",
		}, {
			description: "check the error condition",
			expectedErr: ErrInvalidParameter,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			require := require.New(t)

			mclock := clock.NewMock()

			wm, err := New("name", tc.cfg, UseClock(mclock))

			if tc.expectedErr != nil {
				assert.ErrorIs(err, tc.expectedErr)
				assert.Nil(wm)
				return
			}

			require.NotNil(wm)

			for i := 0; i < tc.pulses; i++ {
				mclock.Add(tc.pulsePeriod)
				wm.Pulse()
			}
			mclock.Add(tc.after)

			assert.Equal(tc.expectListSize, wm.events.Len())
			assert.Equal(tc.total, wm.Total().String())
			assert.Equal(tc.str, wm.String())
			assert.Equal(tc.flow, wm.Flow(tc.over).String())
		})
	}
}
