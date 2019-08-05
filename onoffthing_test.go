// Copyright 2019 Weston Schmidt
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	//"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasics(t *testing.T) {
	assert := assert.New(t)

	opts := OnOffThingOpts{
		Namespace: "testing",
		Name:      "fan"}

	oot := NewOnOffThing(opts)
	s, when := oot.State()
	assert.False(s)
	assert.Equal(time.Time{}, when)

	// Normal, set & wait.
	oot.OnUntil(time.Now().Add(time.Second * 5))
	s, when = oot.State()
	assert.True(s)
	assert.NotEqual(time.Time{}, when)
	time.Sleep(time.Second * 6)
	s, _ = oot.State()
	assert.False(s)

	// Stop it early.
	oot.OnUntil(time.Now().Add(time.Second * 5))
	s, _ = oot.State()
	assert.True(s)
	oot.Off()
	s, _ = oot.State()
	assert.False(s)
	oot.Shutdown()
}

func TestBlackout(t *testing.T) {
	assert := assert.New(t)

	opts := OnOffThingOpts{
		Namespace:      "testing",
		Name:           "blackoutfan",
		BlackoutPeriod: time.Second * 3,
	}

	oot := NewOnOffThing(opts)
	s, _ := oot.State()
	assert.False(s)

	/* Normal, set & wait. */
	oot.OnUntil(time.Now().Add(time.Second * 2))
	s, _ = oot.State()
	assert.True(s)
	time.Sleep(time.Second * 3)
	s, _ = oot.State()
	assert.False(s)

	/* Try to turn it back on during the blackout period */
	oot.OnUntil(time.Now().Add(time.Second * 2))
	s, _ = oot.State()
	assert.False(s)

	time.Sleep(time.Second * 4)

	/* Try to turn it back on after the blackout period */
	s, _ = oot.State()
	assert.False(s)
	oot.OnUntil(time.Now().Add(time.Second * 2))
	s, _ = oot.State()
	assert.True(s)

	/* We're done. */
	oot.Shutdown()
}
