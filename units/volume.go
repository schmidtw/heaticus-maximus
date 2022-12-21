// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package units

import (
	"fmt"
	"strconv"
	"strings"
)

const litersInGallon = 4.54609

// Volume is a measurement of volume stored as a float64 in litres/liters.
type Volume float64

// ParseVolume sets the volume based on the string provided.  Both a number and units
// are required.
func ParseVolume(s string) (Volume, error) {
	list := []struct {
		chompS bool
		suffix string
		volume float64
	}{
		{chompS: true, suffix: "milliliter", volume: 0.001},
		{chompS: true, suffix: "millilitre", volume: 0.001},
		{chompS: true, suffix: "liter", volume: 1.0},
		{chompS: true, suffix: "litre", volume: 1.0},
		{chompS: true, suffix: "gallon", volume: litersInGallon},
		{chompS: false, suffix: "gal", volume: litersInGallon},
		{chompS: false, suffix: "g", volume: litersInGallon},
		{chompS: false, suffix: "ml", volume: 0.001},
		{chompS: false, suffix: "l", volume: 1.0},
	}

	known := make([]string, 0, len(list))

	for _, unit := range list {

		hasSuffix := strings.HasSuffix(strings.ToLower(s), unit.suffix)
		hasS := strings.HasSuffix(strings.ToLower(s), unit.suffix+"s")
		if hasSuffix || unit.chompS && hasS {

			if hasSuffix {
				s = s[:len(s)-len(unit.suffix)]
			} else {
				s = s[:len(s)-len(unit.suffix+"s")]
			}

			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return 0.0, fmt.Errorf("%w: '%s' %v", ErrInvalidUnit, s, err)
			}
			return Volume(n * unit.volume), nil
		}
		known = append(known, unit.suffix)
	}

	return 0.0, fmt.Errorf("%w: unknown unit for '%s' valid: %s", ErrInvalidUnit, s, strings.Join(known, ", "))
}

// Gallons returns the volume as a floating point in gallons.
func (v Volume) Gallons() float64 {
	return float64(v / litersInGallon)
}

// String returns the volume formatted as a string in L.
func (v Volume) String() string {
	return fmt.Sprintf("%.3fL", v)
}
