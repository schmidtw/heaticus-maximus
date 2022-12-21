// SPDX-FileCopyrightText: 2022 Weston Schmidt <weston_schmidt@alumni.purdue.edu>
// SPDX-License-Identifier: Apache-2.0

package main

import "periph.io/x/conn/v3/physic"

type inputs struct {
	i2cFile string
	sample  physic.Frequency
}
