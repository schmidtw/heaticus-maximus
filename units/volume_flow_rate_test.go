// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolumeFlowRate(t *testing.T) {
	tests := []struct {
		in        string
		expect    VolumeFlowRate
		gpm       float64
		str       string
		expectErr error
	}{
		{
			in:     "1gpm",
			expect: VolumeFlowRate(litersInGallon),
			gpm:    1.0,
			str:    "4.546Lpm",
		}, {
			in:     "1.1gpm",
			expect: VolumeFlowRate(litersInGallon),
			gpm:    1.1,
			str:    "5.001Lpm",
		}, {
			in:     "1.0lpm",
			expect: VolumeFlowRate(1.0),
			gpm:    1.0 / litersInGallon,
			str:    "1.000Lpm",
		}, {
			in:        "onelpm", // valid unit, but nonsense number
			expectErr: ErrInvalidUnit,
		}, {
			in:        "1.0", // no units
			expectErr: ErrInvalidUnit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert := assert.New(t)

			vol, err := ParseVolumeFlowRate(tc.in)

			if tc.expectErr == nil {
				assert.NoError(err)
				assert.Equal(tc.str, vol.String())
				assert.Equal(
					fmt.Sprintf("%.6f", tc.gpm),
					fmt.Sprintf("%.6f", vol.GPM()))
				return
			}

			assert.ErrorIs(err, tc.expectErr)
			assert.Equal(VolumeFlowRate(0.0), vol)
		})
	}
}
