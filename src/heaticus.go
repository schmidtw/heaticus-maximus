package main

import (
	"github.com/schmidtw/watermeter"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
	"time"
)

type Heaticus struct {
	Adaptor           *raspi.Adaptor
	LoopMotor         *gpio.RelayDriver
	RecirculatorMotor *gpio.RelayDriver
	UpstairsMotor     *gpio.RelayDriver
	DownstairsMotor   *gpio.RelayDriver
	BathroomMotor     *gpio.RelayDriver
	TowelRackMotor    *gpio.RelayDriver
	ColdInput         *gpio.ButtonDriver
	HotInput          *gpio.ButtonDriver
	LoopInput         *gpio.ButtonDriver
	ColdMeter         watermeter.Watermeter
	HotMeter          watermeter.Watermeter
	LoopMeter         watermeter.Watermeter
}

func (h *Heaticus) evaluate() {

	cold := h.ColdMeter.GetFlow(1 * time.Minute)
	hot := h.HotMeter.GetFlow(5 * time.Minute)
	//loop := h.LoopMeter.GetFlow(1 * time.Minute)

	loopMotor := false
	recirculatorMotor := false
	upstairsMotor := false
	downstairsMotor := false
	bathroomMotor := false
	towelRackMotor := false

	if (0.0 < hot) || (0.0 < cold) {
		loopMotor = true
		recirculatorMotor = true
	}

	/* Check the heater */

	/* Apply settings */
	if loopMotor {
		h.LoopMotor.On()
	} else {
		h.LoopMotor.Off()
	}
	if recirculatorMotor {
		h.RecirculatorMotor.On()
	} else {
		h.RecirculatorMotor.Off()
	}
	if upstairsMotor {
		h.UpstairsMotor.On()
	} else {
		h.UpstairsMotor.Off()
	}
	if downstairsMotor {
		h.DownstairsMotor.On()
	} else {
		h.DownstairsMotor.Off()
	}
	if bathroomMotor {
		h.BathroomMotor.On()
	} else {
		h.BathroomMotor.Off()
	}
	if towelRackMotor {
		h.TowelRackMotor.On()
	} else {
		h.TowelRackMotor.Off()
	}
}

func (h *Heaticus) Init() {
	h.Adaptor = raspi.NewAdaptor()
	h.Adaptor.Connect()

	/* Hook up the Motor Relays */
	h.LoopMotor = gpio.NewRelayDriver(h.Adaptor, "6")
	h.LoopMotor.Start()
	h.LoopMotor.Off()

	h.RecirculatorMotor = gpio.NewRelayDriver(h.Adaptor, "17")
	h.RecirculatorMotor.Start()
	h.RecirculatorMotor.Off()

	h.UpstairsMotor = gpio.NewRelayDriver(h.Adaptor, "8")
	h.UpstairsMotor.Start()
	h.UpstairsMotor.Off()

	h.DownstairsMotor = gpio.NewRelayDriver(h.Adaptor, "9")
	h.DownstairsMotor.Start()
	h.DownstairsMotor.Off()

	h.BathroomMotor = gpio.NewRelayDriver(h.Adaptor, "10")
	h.BathroomMotor.Start()
	h.BathroomMotor.Off()

	h.TowelRackMotor = gpio.NewRelayDriver(h.Adaptor, "11")
	h.TowelRackMotor.Start()
	h.TowelRackMotor.Off()

	/* Hook up the watermeter inputs */
	h.ColdMeter.Init(0.0)
	h.ColdInput = gpio.NewButtonDriver(h.Adaptor, "14")
	h.ColdInput.Start()
	coldInputEvents := h.ColdInput.Subscribe()
	go func() {
		for {
			select {
			case event := <-coldInputEvents:
				if event.Name == gpio.ButtonPush {
					h.ColdMeter.Update(100)
				}
			}
		}
	}()

	h.HotMeter.Init(0.0)
	h.HotInput = gpio.NewButtonDriver(h.Adaptor, "14")
	h.HotInput.Start()
	hotInputEvents := h.HotInput.Subscribe()
	go func() {
		for {
			select {
			case event := <-hotInputEvents:
				if event.Name == gpio.ButtonPush {
					h.HotMeter.Update(100)
				}
			}
		}
	}()

	h.LoopMeter.Init(0.0)
	h.LoopInput = gpio.NewButtonDriver(h.Adaptor, "14")
	h.LoopInput.Start()
	loopInputEvents := h.LoopInput.Subscribe()
	go func() {
		for {
			select {
			case event := <-loopInputEvents:
				if event.Name == gpio.ButtonPush {
					h.LoopMeter.Update(100)
				}
			}
		}
	}()

	/* Hook up the blinking LED */
	go func() {
		led := gpio.NewLedDriver(h.Adaptor, "7")

		for {
			led.Toggle()
			time.Sleep(time.Second)
		}
	}()
}

func (h *Heaticus) Run() {
	for {
		h.evaluate()
		time.Sleep(100 * time.Millisecond)
	}

}

func main() {
	var h Heaticus

	h.Init()

	h.Run()
}
