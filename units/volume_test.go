// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVolume(t *testing.T) {
	tests := []struct {
		in        string
		expect    Volume
		gallons   float64
		str       string
		expectErr error
	}{
		{
			in:      "1G",
			expect:  Volume(litersInGallon),
			gallons: 1.0,
			str:     "4.546L",
		}, {
			in:      "1gal",
			expect:  Volume(litersInGallon),
			gallons: 1.0,
			str:     "4.546L",
		}, {
			in:      "1Gallons",
			expect:  Volume(litersInGallon),
			gallons: 1.0,
			str:     "4.546L",
		}, {
			in:      "1Gallon",
			expect:  Volume(litersInGallon),
			gallons: 1.0,
			str:     "4.546L",
		}, {
			in:      "1.0l",
			expect:  Volume(1.0),
			gallons: 1.0 / litersInGallon,
			str:     "1.000L",
		}, {
			in:      "1.0ml",
			expect:  Volume(0.001),
			gallons: 0.001 / litersInGallon,
			str:     "0.001L",
		}, {
			in:        "onel", // valid unit, but nonsense number
			expectErr: ErrInvalidUnit,
		}, {
			in:        "1.0", // no units
			expectErr: ErrInvalidUnit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			assert := assert.New(t)

			vol, err := ParseVolume(tc.in)

			if tc.expectErr == nil {
				assert.NoError(err)
				assert.Equal(tc.str, vol.String())
				assert.Equal(
					fmt.Sprintf("%.6f", tc.gallons),
					fmt.Sprintf("%.6f", vol.Gallons()))
				return
			}

			assert.ErrorIs(err, tc.expectErr)
			assert.Equal(Volume(0.0), vol)
		})
	}
}
