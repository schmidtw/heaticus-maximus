// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"fmt"
	"strconv"
	"strings"
)

// VolumeFlowRate is a measurement of volume flow rate stored as a float64 in lpm.
type VolumeFlowRate float64

// ParseVolumeFlowRate sets the volume flow rate based on the string provided.
// Both a number and units are required.
func ParseVolumeFlowRate(s string) (VolumeFlowRate, error) {
	list := []struct {
		suffix string
		rate   float64
	}{
		{suffix: "gpm", rate: litersInGallon},
		{suffix: "lpm", rate: 1.0},
	}

	known := make([]string, 0, len(list))

	for _, unit := range list {

		if strings.HasSuffix(strings.ToLower(s), unit.suffix) {
			s = s[:len(s)-len(unit.suffix)]

			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0.0, fmt.Errorf("%w: '%s' %v", ErrInvalidUnit, s, err)
			}
			return VolumeFlowRate(n * unit.rate), nil
		}
		known = append(known, unit.suffix)
	}

	return 0.0, fmt.Errorf("%w: unknown unit for '%s' valid: %s",
		ErrInvalidUnit, s, strings.Join(known, ", "))
}

// Gallons returns the volume flow rate as a floating point in gallons.
func (v VolumeFlowRate) GPM() float64 {
	return float64(v / litersInGallon)
}

// String returns the volume flow rate formatted as a string in L.
func (v VolumeFlowRate) String() string {
	return fmt.Sprintf("%.3fLpm", v)
}
